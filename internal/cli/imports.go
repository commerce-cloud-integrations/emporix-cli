package cli

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

const importScope = "importtool.import_trigger"

const importsUse = "imports"

func importScopes() []string { return []string{importScope} }

func importReader[T any](v view[T]) reader[T] { return reader[T]{view: v, scopes: importScopes()} }

type importConfig struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Enabled   *bool       `json:"enabled"`
	Version   json.Number `json:"version"`
	UpdatedAt string      `json:"updatedAt"`
}

type importStream struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	SourceEntity    string `json:"sourceEntity"`
	FileMode        string `json:"fileMode"`
	TargetType      string `json:"targetType"`
	Mode            string `json:"mode"`
	CompositeParent string `json:"compositeParent"`
	Enabled         *bool  `json:"enabled"`
}

const fanOutMode = "FAN_OUT"

func (s *importStream) source() string {
	switch {
	case s.Mode == fanOutMode && s.CompositeParent != "":
		return spaced("fan-out of", s.CompositeParent)
	case s.FileMode == "":
		return s.SourceEntity
	}
	return spaced(s.SourceEntity, s.FileMode, "files")
}

type importSchedule struct {
	Cron       string `json:"cron"`
	Timezone   string `json:"timezone"`
	Enabled    *bool  `json:"enabled"`
	NextFireAt string `json:"nextFireAt"`
}

type orderedStream struct {
	Stream string   `json:"stream"`
	After  []string `json:"after"`
}

type importedRecord struct {
	ID         string `json:"id"`
	NaturalKey string `json:"naturalKey"`
	TargetType string `json:"targetType"`
	Outcome    string `json:"outcome"`
	ImportedAt string `json:"importedAt"`
}

func newImportsCommand(a *app) *cobra.Command {
	configurations, runsOf := importConfigurations(), configurationRuns()
	streams, streamsOf := importStreams(configurations)
	runs := importRuns(childListLine(configurations, runsOf))
	job := importRunJob(runs)
	reads := slices.Concat([]*cobra.Command{
		newImportConfigurationsCommand(a, configurations, streamsOf, runsOf, importConfigurationWrites(a, configurations,
			job)...),
		newImportStreamsCommand(a, streams),
		newImportRunsCommand(a, runs, importRunWrites(a, runs, job)...),
	}, importDataReads(a), importAnalyticsReads(a))
	return newGroup(importsUse, "Read import configurations, runs and records; queue files, start and schedule runs",
		"Reads the Import Service, which imports master data from external sources into Emporix: "+
			"configurations, their streams (each imports one source, an entity or files, into an Emporix "+
			"type), the file queue of a file-based stream, schedules and runs, the errors and diagnostics of "+
			"a run, the records imported, statistics, job groups, health thresholds and the tenant's import "+
			"limits. Emporix allows these reads only with "+importScope+", with which epx "+
			"also queues files for a stream (and a failed one again), and starts, retries, cancels and schedules runs.",
		reads...)
}

func importConfigurations() resource[importConfig] {
	return resource[importConfig]{
		use:   configurationsUse,
		noun:  "import configurations",
		short: "Read import configurations; start and schedule their runs",
		under: importsUse,
		long: "Reads import configurations, each a set of streams that run together: ENABLED says " +
			"whether it is active, VERSION counts its changes and MODIFIED is when it last changed. Its " +
			"description, source connection, creator and health thresholds are only in `-o json`. The " +
			"service returns the configurations whole, so they are read with one request.",
		path:   "/importtool/{tenant}/configs",
		scopes: importScopes(),
		read:   (*emporix.Client).GetUnpaged,
		view: view[importConfig]{
			singular: "configuration",
			header:   []string{columnID, columnName, columnEnabled, columnVersion, columnModified},
			row: func(c importConfig, _ language) []string {
				return []string{c.ID, c.Name, formatBool(c.Enabled), c.Version.String(), c.UpdatedAt}
			},
		},
	}
}

func newImportConfigurationsCommand(a *app, configurations resource[importConfig], streams child[importStream],
	runs child[importRun], writes ...*cobra.Command) *cobra.Command {
	return newResourceCommand(a, configurations, append([]*cobra.Command{
		childCommand(a, configurations, streams),
		childCommand(a, configurations, child[importSchedule]{
			use:   "schedule",
			short: "Show the run schedule of a configuration",
			long: "Reads when an import configuration runs by itself: CRON is a cron expression with six fields " +
				"(seconds first), evaluated in TIME ZONE, and NEXT RUN when it fires next. A configuration " +
				"without a schedule is not found (exit 4).",
			fetch: single{items: oneItem},
			view: view[importSchedule]{
				singular: "schedule",
				header:   []string{"CRON", "TIME ZONE", columnEnabled, "NEXT RUN"},
				row: func(s importSchedule, _ language) []string {
					return []string{s.Cron, s.Timezone, formatBool(s.Enabled), s.NextFireAt}
				},
			},
		}),
		childCommand(a, configurations, child[orderedStream]{
			use:   "stream-order",
			short: "Show the order the streams of a configuration run in",
			long: "Reads the order in which the streams of an import configuration run, which the service " +
				"derives from their dependencies (also those their mappings add, and that of a FAN_OUT stream on " +
				"the stream it fans out from), their seq breaking ties: each STREAM runs after the streams that " +
				"AFTER names.",
			fetch:   single{items: streamOrderItems},
			unknown: listsNothing,
			view: view[orderedStream]{
				singular: "stream order",
				header:   []string{columnStream, "AFTER"},
				row: func(s orderedStream, _ language) []string {
					return []string{s.Stream, strings.Join(s.After, ", ")}
				},
			},
		}),
		childCommand(a, configurations, runs),
	}, writes...)...)
}

func configurationStreams(v view[importStream]) child[importStream] {
	return child[importStream]{
		use:   streamsUse,
		short: "List the streams of a configuration",
		long: "Reads the streams of an import configuration, in the order of their seq. The service returns " +
			"them whole, so they are read with one request.",
		fetch:   listing{read: (*emporix.Client).GetUnpaged},
		unknown: listsNothing,
		view:    v,
	}
}

const streamsUse = "streams"

func importStreams(configurations resource[importConfig]) (resource[importStream], child[importStream]) {
	queue := "`epx " + strings.Join([]string{importsUse, streamsUse, filesUse}, " ") + "`"
	v := importStreamView(queue)
	streamsOf := configurationStreams(v)
	listedBy := childListLine(configurations, streamsOf)
	long := "Reads an import stream, which imports one source of a configuration, an entity or files, " +
		"into an Emporix type. A stream's ID is the one that `epx " + listedBy + "` shows."
	return resource[importStream]{
		use:      streamsUse,
		noun:     "import streams",
		short:    "Read import streams and the records they imported; queue their files",
		under:    importsUse,
		listedBy: listedBy,
		long:     long,
		about:    joinSentences(long, queue+" reads and changes the queue of the files that a file-based stream imports."),
		path:     "/importtool/{tenant}/streams",
		scopes:   importScopes(),
		view:     v,
	}, streamsOf
}

func importStreamView(queue string) view[importStream] {
	return view[importStream]{
		singular: "stream",
		header:   []string{columnID, columnName, "SOURCE", "TARGET TYPE", "MODE", columnEnabled},
		legend: "SOURCE is the stream's source entity, followed for a file-based stream (which may name no " +
			"entity) by SNAPSHOT files or INCREMENTAL files, read as " + queue + " explains; a FAN_OUT stream " +
			"has no source of its own, so SOURCE names the stream it fans out from. " +
			"TARGET TYPE is the custom entity type that a stream writing custom entities (targetWriter " +
			"emporix.customEntity) writes to (a discriminator, in `-o json`, may route records to others); " +
			"for the other writers (products, categories, prices, stock, media and category assignments) it " +
			"is a label that groups the records the stream imported: PRODUCT for products and CATEGORY for " +
			"categories by default, set by the service for a FAN_OUT stream. MODE is STANDALONE (a stream " +
			"that writes its own objects), COMPOSITE_CHILD or COMPOSITE_MERGE for one that adds to the records " +
			"of another, " +
			"or FAN_OUT for one that writes further objects, such as a product's prices, for each record of " +
			"the stream it fans out from. A stream's writer (targetWriter), position (seq), fan-out settings " +
			"(fanOut, fanOutPath, fanOutKey), target options (targetOptions), source dataset (sourceDataset, " +
			"such as the key column and the file format), whether it takes each file as a snapshot " +
			"(snapshotFiles), connection, and delta, delete, write and health settings are only in `-o json`.",
		row: func(s importStream, _ language) []string {
			return []string{s.ID, s.Name, s.source(), s.TargetType, s.Mode, formatBool(s.Enabled)}
		},
	}
}

func newImportStreamsCommand(a *app, streams resource[importStream]) *cobra.Command {
	return newGroup(streams.use, streams.summary(), streams.groupHelp(),
		streams.getCommand(a),
		importReader(importedRecordView()).command(a, &command{
			use:   "records " + streams.argument(),
			short: "List the records a stream imported",
			long: "Reads the records a stream imported, with the columns of `epx imports records`. " +
				streams.argumentSentence(listsNothing),
			req:   request{path: "/importtool/{tenant}/data/streams/" + placeholderID + "/records"},
			fetch: listing{read: (*emporix.Client).GetAllContentPages},
		}),
		newImportFilesCommand(a, streams))
}

func importedRecordView() view[importedRecord] {
	return view[importedRecord]{
		singular: "imported record",
		header:   []string{columnID, columnKey, columnType, "OUTCOME", "IMPORTED"},
		row: func(r importedRecord, _ language) []string {
			return []string{r.ID, r.NaturalKey, r.TargetType, r.Outcome, r.ImportedAt}
		},
	}
}

func importDataReads(a *app) []*cobra.Command {
	return []*cobra.Command{
		importReader(stringsView("imported type", columnType)).command(a, &command{
			use:   typesUse,
			short: "List the types that hold imported records",
			long: "Reads the target types that hold imported records (`epx imports records` lists those of " +
				"one). The service returns them whole, so they are read with one request.",
			req:   request{path: "/importtool/{tenant}/data/types"},
			fetch: listing{read: (*emporix.Client).GetUnpaged},
		}),
		importReader(importedRecordView()).command(a, &command{
			use:   "records TYPE",
			short: "List the imported records of a type",
			long: "Reads the records imported into a target type: KEY is a record's natural key in the source, " +
				"OUTCOME how it was last imported (UPSERTED, DELETED at the source, FAILED or DRY_RUN; older " +
				"records may show DELETED_TARGET) and IMPORTED when. `-o json` adds its stored fields, the " +
				"Emporix id it was written under, and its configuration, stream and run. TYPE is one that " +
				"`epx imports types` shows; an unknown one lists nothing. To find the records whose natural " +
				"key (with or without its key name, e.g. ID:SKU-1001; case-sensitive) or Emporix id (in any case) " +
				"starts with a text, or those of one outcome (UPSERTED, DELETED, FAILED or DRY_RUN, in upper " +
				"case), one page at a time: " +
				"`epx api get /importtool/{tenant}/data/records -p type=TYPE -p search=TEXT -p outcome=FAILED " +
				"--scope " + importScope + "`.",
			req:   request{path: "/importtool/{tenant}/data/records", query: url.Values{typeParam: {placeholderID}}},
			fetch: listing{read: (*emporix.Client).GetAllContentPages},
		}),
	}
}

func streamOrderItems(body json.RawMessage) ([]json.RawMessage, error) {
	var order *struct {
		Order   []string            `json:"order"`
		Prereqs map[string][]string `json:"prereqs"`
	}
	if err := json.Unmarshal(body, &order); err != nil || order == nil || order.Order == nil {
		return nil, errors.New("expected an object with the streams in order and their prerequisites")
	}
	streams := make([]orderedStream, len(order.Order))
	for i, name := range order.Order {
		streams[i] = orderedStream{Stream: name, After: order.Prereqs[name]}
	}
	return marshalEach(streams)
}
