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

// importScope is the one scope that reads the Import Service. It also starts,
// schedules and cancels import runs, so every read of the service is opt-in.
const importScope = "importtool.import_trigger"

// importsUse is the name of the group of Import Service reads.
const importsUse = "imports"

// importScopes returns the scopes of an Import Service read.
func importScopes() []string { return []string{importScope} }

// importReader returns the reader of an Import Service read whose items v
// shows.
func importReader[T any](v view[T]) reader[T] { return reader[T]{view: v, scopes: importScopes()} }

// importConfig is the part of an import configuration shown in tables.
type importConfig struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Enabled   *bool       `json:"enabled"`
	Version   json.Number `json:"version"` // kept as sent
	UpdatedAt string      `json:"updatedAt"`
}

// importStream is the part of an import stream shown in tables.
type importStream struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SourceEntity string `json:"sourceEntity"`
	FileMode     string `json:"fileMode"` // SNAPSHOT or INCREMENTAL, for a file-based stream only
	TargetType   string `json:"targetType"`
	Mode         string `json:"mode"`
	Enabled      *bool  `json:"enabled"`
}

// source returns what s imports: its source entity, followed for a
// file-based stream by how it reads its files, e.g. "SNAPSHOT files", as such
// a stream may name no entity.
func (s *importStream) source() string {
	if s.FileMode == "" {
		return s.SourceEntity
	}
	return spaced(s.SourceEntity, s.FileMode, "files")
}

// importSchedule is the part of a configuration's run schedule shown in tables.
type importSchedule struct {
	Cron       string `json:"cron"`
	Timezone   string `json:"timezone"`
	Enabled    *bool  `json:"enabled"`
	NextFireAt string `json:"nextFireAt"`
}

// orderedStream is a stream of a configuration, in the order the streams run
// in, with the streams that must run before it (streamOrderItems).
type orderedStream struct {
	Stream string   `json:"stream"`
	After  []string `json:"after"`
}

// importedRecord is the part of an imported record shown in tables.
type importedRecord struct {
	ID         string `json:"id"`
	NaturalKey string `json:"naturalKey"`
	TargetType string `json:"targetType"`
	Outcome    string `json:"outcome"`
	ImportedAt string `json:"importedAt"`
}

// newImportsCommand reads the Import Service: its configurations, streams
// and their file queues, runs, imported records and statistics. Emporix reads
// them only with the write-capable importtool.import_trigger scope, so every
// read is opt-in. With that scope, it also queues files for a stream
// (newImportFilesCommand), starts, retries and cancels runs and schedules
// them (importConfigurationWrites, importRunWrites).
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
			"also queues files for a stream, and starts, retries, cancels and schedules runs.",
		reads...)
}

// importConfigurations returns import configurations.
func importConfigurations() resource[importConfig] {
	return resource[importConfig]{
		use:   configurationsUse,
		noun:  "import configurations",
		short: "Read import configurations; start and schedule their runs",
		under: importsUse,
		long: "Reads import configurations, each a set of streams that run together: ENABLED says " +
			"whether it is active, VERSION counts its changes and MODIFIED is when it last changed. Its " +
			"description, source connection, creator and health thresholds are only in -o json. The " +
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

// newImportConfigurationsCommand reads import configurations with their
// streams, schedule, stream order and runs, followed by writes.
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
				"derives from their dependencies (also those their mappings add) and sequence: each STREAM " +
				"runs after the streams that AFTER names.",
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

// configurationStreams declares the streams of an import configuration,
// shown in v, the table of streams.
func configurationStreams(v view[importStream]) child[importStream] {
	return child[importStream]{
		use:   streamsUse,
		short: "List the streams of a configuration",
		long: "Reads the streams of an import configuration, in their sequence. The service returns them " +
			"whole, so they are read with one request.",
		fetch:   listing{read: (*emporix.Client).GetUnpaged},
		unknown: listsNothing,
		view:    v,
	}
}

// streamsUse names import streams, and the streams of a configuration.
const streamsUse = "streams"

// importStreams returns import streams, which the service reads one at a time
// by their own id, and the streams of a configuration of configurations
// (configurationStreams), which list them, both with the table of streams,
// which names the command of their file queue (queue).
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

// importStreamView is the table of import streams, whose legend names queue,
// the command of their file queue, as help quotes it.
func importStreamView(queue string) view[importStream] {
	return view[importStream]{
		singular: "stream",
		header:   []string{columnID, columnName, "SOURCE", "TARGET TYPE", "MODE", columnEnabled},
		legend: "SOURCE is the stream's source entity, followed for a file-based stream (which may name no " +
			"entity) by SNAPSHOT files or INCREMENTAL files, read as " + queue + " explains. " +
			"TARGET TYPE is the Emporix type it writes to (a discriminator, in -o json, may route records " +
			"to others), MODE STANDALONE, or COMPOSITE_CHILD or COMPOSITE_MERGE for a stream that adds to " +
			"the records of another. A stream's source dataset (sourceDataset, such as the key column and " +
			"the file format), connection, and delta, delete, write and health settings are only in " +
			"-o json.",
		row: func(s importStream, _ language) []string {
			return []string{s.ID, s.Name, s.source(), s.TargetType, s.Mode, formatBool(s.Enabled)}
		},
	}
}

// newImportStreamsCommand reads one import stream and the records a stream
// imported, followed by the group of a stream's file queue.
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

// importedRecordView is the table of imported records.
func importedRecordView() view[importedRecord] {
	return view[importedRecord]{
		singular: "imported record",
		header:   []string{columnID, columnKey, columnType, "OUTCOME", "IMPORTED"},
		row: func(r importedRecord, _ language) []string {
			return []string{r.ID, r.NaturalKey, r.TargetType, r.Outcome, r.ImportedAt}
		},
	}
}

// importDataReads returns the reads of the imported records: the types that
// hold them and the records of one type.
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
				"OUTCOME how it was last imported (UPSERTED, DELETED at the source, DELETED_TARGET in " +
				"Emporix, FAILED or DRY_RUN) and IMPORTED when. -o json adds its stored fields, the Emporix " +
				"id it was written under, and its configuration, stream and run. TYPE is one that " +
				"`epx imports types` shows; an unknown one lists nothing. To search keys or filter by " +
				"outcome, one page at a time: `epx api get /importtool/{tenant}/data/records -p type=TYPE " +
				"-p search=TEXT -p outcome=FAILED --scope " + importScope + "`.",
			req:   request{path: "/importtool/{tenant}/data/records", query: url.Values{typeParam: {placeholderID}}},
			fetch: listing{read: (*emporix.Client).GetAllContentPages},
		}),
	}
}

// streamOrderItems returns the streams of a stream order response
// ({"order": [names], "prereqs": {name: [names]}}) in the order they run, each
// with the streams that must run before it (orderedStream).
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
