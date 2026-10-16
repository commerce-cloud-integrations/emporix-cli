package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// importRun is the part of an import run shown in tables.
type importRun struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"`
	Mode        string      `json:"mode"`
	DryRun      bool        `json:"dryRun"`
	StartedAt   string      `json:"startedAt"`
	FinishedAt  string      `json:"finishedAt"`
	RecordsRead json.Number `json:"recordsRead"`
	Failed      json.Number `json:"failed"`
	Message     string      `json:"message"` // why it failed, where it says; only in -o json
}

// mode returns the run's mode, FULL or DELTA, followed by "dry run" for a run
// that wrote nothing.
func (r *importRun) mode() string {
	if r.DryRun {
		return spaced(r.Mode, "dry run")
	}
	return r.Mode
}

// runStream is the progress of one stream in an import run.
type runStream struct {
	StreamName  string      `json:"streamName"`
	Status      string      `json:"status"`
	RecordsRead json.Number `json:"recordsRead"`
	Created     json.Number `json:"created"`
	Updated     json.Number `json:"updated"`
	Skipped     json.Number `json:"skipped"`
	Deleted     json.Number `json:"deleted"`
	Failed      json.Number `json:"failed"`
	Message     string      `json:"message"`
}

// importError is the part of an error recorded during an import run shown in
// tables.
type importError struct {
	ID         string `json:"id"`
	NaturalKey string `json:"naturalKey"`
	Stage      string `json:"stage"`
	ErrorCode  string `json:"errorCode"`
	Message    string `json:"message"`
}

// diagnosticRow is a source row behind a run's feed-quality counters.
type diagnosticRow struct {
	Kind         string      `json:"kind"`
	StreamName   string      `json:"streamName"`
	NaturalKey   string      `json:"naturalKey"`
	LinkingField string      `json:"linkingField"`
	LinkingValue string      `json:"linkingValue"`
	WaitingLines json.Number `json:"waitingLines"`
	Occurrence   json.Number `json:"occurrence"`
}

// maxDiagnosticRows is the most diagnostic rows the Import Service returns at
// once; diagnostics asks for that many and fails when a run recorded more.
const maxDiagnosticRows = 50000

// configurationRuns declares the runs of an import configuration.
func configurationRuns() child[importRun] {
	return child[importRun]{
		use:     "runs",
		short:   "List the runs of a configuration",
		long:    "Reads the run history of an import configuration, most recent first.",
		fetch:   listing{read: (*emporix.Client).GetAllContentPages},
		unknown: listsNothing,
		view:    importRunView(),
	}
}

// importRuns returns import runs, which the service reads one at a time by
// id; the command line listedBy lists them (the runs of a configuration).
func importRuns(listedBy string) resource[importRun] {
	return resource[importRun]{
		use:      "runs",
		noun:     "import runs",
		short:    "Read, cancel and retry import runs",
		under:    importsUse,
		listedBy: listedBy,
		long: "Reads import runs; a run's ID is the one that `epx " + listedBy + "` shows.\n" +
			"Its -o json adds the progress of each stream of the run, which `epx imports runs streams` shows.",
		path:   "/importtool/{tenant}/runs",
		item:   detailRun, // the run of its details
		scopes: importScopes(),
		view:   importRunView(),
	}
}

// importRunView is the table of import runs.
func importRunView() view[importRun] {
	return view[importRun]{
		singular: "run",
		header:   []string{columnID, columnStatus, "MODE", "STARTED", "FINISHED", "READ", columnFailed},
		legend: "STATUS is QUEUED, RUNNING, RESUMABLE (interrupted; it resumes by itself), SUCCEEDED,\n" +
			"PARTIAL (records failed or a limit skipped some, or a stream did not run), FAILED, CANCELLED\n" +
			"or ABORTED (refused: nothing was read), MODE FULL or DELTA, followed by dry run for a run that\n" +
			"wrote nothing, STARTED when the run was requested, FINISHED when it ended, READ the records it\n" +
			"read and FAILED those that failed. -o json adds its other counters, message and queue position.",
		row: func(r importRun, _ language) []string {
			return []string{
				r.ID, r.Status, r.mode(), r.StartedAt, r.FinishedAt, r.RecordsRead.String(), r.Failed.String(),
			}
		},
	}
}

// newImportRunsCommand reads one import run, the progress of its streams, its
// errors and its diagnostics, followed by writes.
func newImportRunsCommand(a *app, runs resource[importRun], writes ...*cobra.Command) *cobra.Command {
	limit := strconv.Itoa(maxDiagnosticRows)
	return newGroup(runs.use, runs.summary(), runs.long, append([]*cobra.Command{
		runs.getCommand(a),
		importReader(runStreamView()).command(a, &command{
			use:   "streams " + runs.argument(),
			short: "Show the progress of each stream of a run",
			long: "Reads the progress of each stream of an import run: STATUS is QUEUED, RUNNING, SUCCEEDED,\n" +
				"PARTIAL, FAILED, SKIPPED or ABORTED (not run because of its configuration, such as mappings\n" +
				"never published, as MESSAGE says), then the records the stream read, created, updated,\n" +
				"skipped as unchanged, deleted and failed to import. A queued run lists nothing yet.\n" +
				runs.argumentSentence(notFound),
			req:   request{path: runs.path + "/" + placeholderID},
			fetch: single{items: detailStreams},
		}),
		childCommand(a, runs, child[importError]{
			use:   "errors",
			short: "List the errors of a run",
			long: "Reads the errors recorded during an import run: KEY is the natural key of the record that\n" +
				"failed, STAGE the step it failed in (such as MAPPING or UPSERT), CODE the error code and\n" +
				"MESSAGE what went wrong. Its stream and time are only in -o json.",
			fetch:   listing{read: (*emporix.Client).GetAllContentPages},
			unknown: listsNothing,
			view: view[importError]{
				singular: "run error",
				header:   []string{columnID, columnKey, "STAGE", columnCode, columnMessage},
				long:     []string{columnMessage},
				row: func(e importError, _ language) []string {
					return []string{e.ID, e.NaturalKey, e.Stage, e.ErrorCode, e.Message}
				},
			},
		}),
		childCommand(a, runs, child[diagnosticRow]{
			use:   "diagnostics",
			short: "List the source rows behind a run's repeated keys and missing parents",
			long: "Reads the source rows behind a run's feed-quality counters, which do not fail a run: KIND is\n" +
				"REPEATED_KEY for a row whose KEY the run had already imported (OCCURRENCE times before), or\n" +
				"UNRESOLVED_PARENT for a child line whose parent was missing: LINKING FIELD and LINKING VALUE\n" +
				"are the parent key it points at, WAITING LINES the lines waiting for that parent. The service\n" +
				"keeps a capped sample per stream and kind; -o json says how many rows it recorded, its cap\n" +
				"and whether a stream reached it (the run's counters give the totals). Every recorded row is\n" +
				"read, up to " + limit + "; a run that recorded more fails, and its rows are read per stream:\n" +
				"`epx api get /importtool/{tenant}/runs/RUN_ID/diagnostics -p streamId=ID -p limit=" + limit + "\n" +
				"--scope " + importScope + "`.",
			query:   url.Values{"limit": {limit}},
			fetch:   single{items: diagnosticRows},
			unknown: listsNothing,
			view:    diagnosticView(),
		}),
	}, writes...)...)
}

// runStreamView is the table of the progress of each stream of a run.
func runStreamView() view[runStream] {
	return view[runStream]{
		singular: "run stream",
		header: []string{
			columnStream, columnStatus, "READ", columnCreated, "UPDATED", "SKIPPED", "DELETED", columnFailed, columnMessage,
		},
		long: []string{columnMessage},
		row: func(s runStream, _ language) []string {
			return []string{s.StreamName, s.Status, s.RecordsRead.String(), s.Created.String(), s.Updated.String(),
				s.Skipped.String(), s.Deleted.String(), s.Failed.String(), s.Message}
		},
	}
}

// diagnosticView is the table of the diagnostic rows of a run.
func diagnosticView() view[diagnosticRow] {
	return view[diagnosticRow]{
		singular: "run diagnostics",
		header: []string{
			"KIND", columnStream, columnKey, "LINKING FIELD", "LINKING VALUE", "WAITING LINES", "OCCURRENCE",
		},
		row: func(d diagnosticRow, _ language) []string {
			return []string{d.Kind, d.StreamName, d.NaturalKey, d.LinkingField, d.LinkingValue, d.WaitingLines.String(),
				d.Occurrence.String()}
		},
	}
}

// runDetail is a run with the progress of its streams, as the Import Service
// reads one run.
type runDetail struct {
	Run     json.RawMessage   `json:"run"`
	Streams []json.RawMessage `json:"streams"` // none for a queued run
}

// decodeRunDetail decodes a run's details, which must hold the run as one
// item (oneItem): details whose run is null, empty or not an object are
// malformed, also where only the streams are read.
func decodeRunDetail(body json.RawMessage) (runDetail, error) {
	var detail runDetail
	err := json.Unmarshal(body, &detail)
	if err == nil {
		_, err = oneItem(detail.Run)
	}
	if err != nil {
		return runDetail{}, errors.New("expected an object with a run and its streams")
	}
	return detail, nil
}

// detailRun returns the run of a run's details.
func detailRun(body json.RawMessage) ([]json.RawMessage, error) {
	detail, err := decodeRunDetail(body)
	if err != nil {
		return nil, err
	}
	return []json.RawMessage{detail.Run}, nil
}

// detailStreams returns the progress of each stream of a run's details.
func detailStreams(body json.RawMessage) ([]json.RawMessage, error) {
	detail, err := decodeRunDetail(body)
	return detail.Streams, err
}

// diagnosticRows returns the rows of a run's diagnostics, failing when the
// service recorded more rows than it returned (a run that recorded more than
// maxDiagnosticRows), as the rows would then be a part only.
func diagnosticRows(body json.RawMessage) ([]json.RawMessage, error) {
	var diagnostics struct {
		Rows     []json.RawMessage `json:"rows"`
		Recorded *int              `json:"recorded"` // the rows recorded for the run; unchecked when left out
	}
	if err := json.Unmarshal(body, &diagnostics); err != nil || diagnostics.Rows == nil {
		return nil, errors.New("expected an object with the diagnostic rows and an integer recorded, if any")
	}
	if recorded := diagnostics.Recorded; recorded != nil && *recorded > len(diagnostics.Rows) {
		return nil, fmt.Errorf("the run recorded %d rows, more than the %d returned; read them per stream, "+
			"as the help says", *recorded, len(diagnostics.Rows))
	}
	return diagnostics.Rows, nil
}
