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

// runStream is the progress of one stream in an import run: the records it
// read, which its other counters add up to, and its message.
type runStream struct {
	StreamName        string      `json:"streamName"`
	Status            string      `json:"status"`
	RecordsRead       json.Number `json:"recordsRead"`
	Created           json.Number `json:"created"`
	Updated           json.Number `json:"updated"`
	Skipped           json.Number `json:"skipped"`
	Deleted           json.Number `json:"deleted"`
	Failed            json.Number `json:"failed"`
	UnresolvedParents json.Number `json:"unresolvedParents"`
	Collapsed         json.Number `json:"collapsed"`
	Embedded          json.Number `json:"embedded"`
	Message           string      `json:"message"`
	Sweep             bool        `json:"sweep"` // the repair pass of a stream, not a second stream
}

// stream returns the name of the stream of s, followed by "sweep" for the
// repair pass that read again the parents a child stream had not found; a
// progress without a name has no identifier, also when it is a sweep.
func (s *runStream) stream() string {
	if s.Sweep && s.StreamName != "" {
		return spaced(s.StreamName, "sweep")
	}
	return s.StreamName
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
		use:   "runs",
		short: "List the runs of a configuration",
		long: "Reads the run history of an import configuration, most recent first; `epx imports runs streams` " +
			"shows the counters of each stream of a run.",
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
		long: "Reads import runs; a run's ID is the one that `epx " + listedBy + "` shows. " +
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
		header:   []string{columnID, columnStatus, "MODE", "STARTED", "FINISHED", columnRead, columnFailed},
		legend: "STATUS is QUEUED, RUNNING, RESUMABLE (interrupted; it resumes by itself), SUCCEEDED, " +
			"PARTIAL (records failed or a limit skipped some, or a stream did not run), FAILED, CANCELLED " +
			"or ABORTED (refused: nothing was read), MODE FULL or DELTA, followed by dry run for a run that " +
			"wrote nothing, STARTED when the run was requested, FINISHED when it ended, READ the records it " +
			"read and FAILED those that failed. -o json adds its message, queue position and other counters; " +
			"READ is created + updated + skipped + deleted + failed + unresolvedParents + collapsed + " +
			"embedded.",
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
			long: "Reads the progress of each stream of an import run: STATUS is QUEUED, RUNNING, SUCCEEDED, " +
				"PARTIAL, FAILED, SKIPPED or ABORTED (not run because of its configuration, such as mappings " +
				"never published, as MESSAGE says). READ counts the records the stream read, and the counters " +
				"after it add up to READ: CREATED, UPDATED, SKIPPED (unchanged, or not written because the " +
				"tenant's limit of records per type was reached), DELETED, FAILED, UNRESOLVED (child records " +
				"whose parent was not found; `epx imports runs diagnostics` lists those parents), COLLAPSED " +
				"(rows that repeated a key of the same batch, merged before the write, the last one winning) " +
				"and EMBEDDED (child records that their parent wrote inside its own record). A STREAM followed " +
				"by sweep is the repair pass that read again the parents a child stream had not found, not a " +
				"second stream. -o json adds duplicateKeys (rows whose key the run had imported already, which " +
				"`epx imports runs diagnostics` lists too), capSkipped (whether the tenant's limit of records " +
				"per type skipped records), keylessSeen (whether a record had no key: an error NO_KEY) and the " +
				"files a file-based stream read (filesRead, fileNames). A queued run lists nothing yet. " +
				runs.argumentSentence(notFound),
			req:   request{path: runs.path + "/" + placeholderID},
			fetch: single{items: detailStreams},
		}),
		childCommand(a, runs, runErrors()),
		childCommand(a, runs, child[diagnosticRow]{
			use:   "diagnostics",
			short: "List the source rows behind a run's repeated keys and missing parents",
			long: "Reads the source rows behind a run's feed-quality counters, which do not fail a run: KIND is " +
				"REPEATED_KEY for a row whose KEY the run had already imported (OCCURRENCE times before), or " +
				"UNRESOLVED_PARENT for a child line whose parent was missing: LINKING FIELD and LINKING VALUE " +
				"are the parent key it points at, WAITING LINES the lines waiting for that parent. The service " +
				"keeps a capped sample per stream and kind; -o json says how many rows it recorded, its cap " +
				"and whether a stream reached it (the run's counters give the totals). Every recorded row is " +
				"read, up to " + limit + "; a run that recorded more fails, and its rows are read per stream: " +
				"`epx api get /importtool/{tenant}/runs/RUN_ID/diagnostics -p streamId=ID -p limit=" + limit + " " +
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
			columnStream, columnStatus, columnRead, columnCreated, "UPDATED", "SKIPPED", "DELETED", columnFailed,
			"UNRESOLVED", "COLLAPSED", "EMBEDDED", columnMessage,
		},
		long: []string{columnMessage},
		row: func(s runStream, _ language) []string {
			return []string{s.stream(), s.Status, s.RecordsRead.String(), s.Created.String(), s.Updated.String(),
				s.Skipped.String(), s.Deleted.String(), s.Failed.String(), s.UnresolvedParents.String(),
				s.Collapsed.String(), s.Embedded.String(), s.Message}
		},
	}
}

// runErrors declares the errors recorded during an import run.
func runErrors() child[importError] {
	return child[importError]{
		use:   "errors",
		short: "List the errors of a run",
		long: "Reads the errors recorded during an import run: KEY is the natural key of the record that " +
			"failed, STAGE the step it failed in (such as MAPPING or UPSERT), CODE what kind of error it is " +
			"and MESSAGE what went wrong. Its stream and time are only in -o json. CODE is VALIDATION (the " +
			"mapped record was invalid, so it was not written), NO_KEY (its key column was empty), UPSERT " +
			"(the target refused the write), CONFIGURATION (something to change in the import's own " +
			"configuration, such as a missing base URL, refused credentials or an unknown path), " +
			"EXTERNAL_SYSTEM (another system answered with an error, such as the source with a 5xx), MERGE " +
			"or PATCH_EMBED (a composite merge or an embed into the parent failed), NO_LINKING_VALUE (a " +
			"child row without the key of its parent), MARKER_DELETE, TOMBSTONE_DELETE or TOMBSTONE_NO_KEY " +
			"(a delete the source marked could not be applied), REJECTED_BY_SOURCE (the source refused the " +
			"record for its data, such as a file row with more array elements than the stream allows; the " +
			"other records are imported) or STREAM_ERROR (the stream failed as a whole). For CONFIGURATION " +
			"and EXTERNAL_SYSTEM, MESSAGE says what to change.",
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
