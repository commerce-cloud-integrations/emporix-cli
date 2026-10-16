package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// perTypeWrites declares the writes of one item of a Search Service
// collection kept per custom entity type (perTypeCommand): update TYPE ID,
// an upsert of the JSON object in --file, whose localized text is sent
// as maps only (mapsOnly), and delete TYPE ID, both with searchManageScope.
// Each text field says what its write sends and what Emporix answers; what
// TYPE and ID are, how it is sent and the scope sentence follow.
type perTypeWrites struct {
	update  string   // what update sends
	example string   // a command line of update, indented by two spaces
	delete  string   // what delete sends
	answer  answerer // how update and delete end: by following the job that they start; by printing when nil
}

// commands returns update and delete of one item of a collection, which
// singular names, sent to item; arguments says what TYPE and ID are, and ids
// lists them for completion.
func (p perTypeWrites) commands(a *app, singular string, item request, arguments string,
	ids argumentIDs) []*cobra.Command {
	manage := []string{searchManageScope}
	return []*cobra.Command{
		(&write{
			use: "update TYPE ID", short: "Upsert a " + singular, method: http.MethodPut,
			long: p.update + "\n" + arguments + ", or a new ID.", example: p.example,
			req: item, scopes: manage, body: &payload{}, language: mapsOnly, ids: ids, answer: p.answer,
		}).command(a),
		(&write{
			use: "delete TYPE ID", short: "Delete a " + singular, method: http.MethodDelete,
			long: p.delete + "\n" + arguments + ".",
			req:  item, scopes: manage, body: noBody{}, ids: ids, answer: p.answer,
		}).command(a),
	}
}

// indexJobWait returns how --wait follows an index job of jobs that a write
// of search indexes starts: it reads the job as `search jobs get` does until
// it is no longer in_progress, and succeeds only when it ended with success.
// Each write names the jobs that it starts by their IDs alone (jobIDs).
func indexJobWait(jobs resource[indexJob]) *jobWait {
	return followJob(jobs, (*indexJob).state, []string{"in_progress"}, "success")
}

// state returns what --wait follows of j: its response, what it recorded
// when it ended, which says why it failed.
func (j *indexJob) state() jobState {
	return jobState{ID: j.ID, Status: j.Status, Failure: j.Response}
}

// searchIndexWrites declares the writes of one search index, which start the
// job that builds or removes it, which job follows: Emporix answers with the
// job's id alone, or with nothing where it starts none.
func searchIndexWrites(job *jobWait) perTypeWrites {
	return perTypeWrites{
		update: "Creates the search index ID of the custom entity type TYPE, or replaces it (an upsert),\n" +
			"with the one in --file, a JSON object with its fields (each with its path and how it is\n" +
			"searched) and optionally name and description. A job builds a new index, and one whose\n" +
			"fields changed or that is building or failed: Emporix then answers 202 Accepted with the\n" +
			"job's id (`epx search jobs get ID` reads the job), else 204 No Content. With\n" +
			"metadata.version, as `epx search indexes get TYPE ID -o json` prints it, Emporix replaces\n" +
			"only that version; another, or a job of the index that is still running, is a conflict\n" +
			"(exit 5).",
		example: "  epx search indexes update vehicle vehicles --file index.json --dry-run\n" +
			"  epx search indexes update vehicle vehicles --file index.json --wait --yes",
		delete: "Starts the job that deletes the search index ID of the custom entity type TYPE: Emporix\n" +
			"answers with the job's id (`epx search jobs get ID` reads the job), and with 409 Conflict\n" +
			"(exit 5) while another job of the index is running.",
		answer: job.answering(jobIDs{field: "id"}),
	}
}

// savedSearchWrites declares the writes of one saved search.
func savedSearchWrites() perTypeWrites {
	return perTypeWrites{
		update: "Creates the saved search ID of the custom entity type TYPE, or replaces it (an upsert),\n" +
			"with the one in --file, a JSON object with index (the ID of a search index of TYPE),\n" +
			"queries, filters or both, and optionally name and description. Emporix answers a new\n" +
			"one with its id (201 Created), and a replaced one with 204 No Content. Replacing one\n" +
			"needs its metadata.version, as `epx search queries get TYPE ID -o json` prints it; a\n" +
			"stale one is a conflict (exit 5). An unknown index is not found (exit 4).",
		example: "  epx search queries update vehicle by-make --file by-make.json --dry-run",
		delete: "Deletes the saved search ID of the custom entity type TYPE; an unknown one is not found\n" +
			"(exit 4).",
	}
}

// searchIndexImport returns import, which creates or updates each search
// index of an export package, as update does, starting the jobs that job
// follows, which Emporix names by their IDs alone, in its result per index.
func searchIndexImport(a *app, indexes resource[searchIndex], job *jobWait) *cobra.Command {
	return (&write{
		use: importUse, short: "Import search indexes", method: http.MethodPost,
		long: "Creates or updates each search index of the export package in --file, as update does. The\n" +
			"package is the whole response of `epx search indexes export -o json`, with its exportedAt\n" +
			"(Emporix answers 400 without it). Emporix answers with a\n" +
			"result per index, in the order of the package, with the id of the job that builds it\n" +
			"where one started. When an index fails, those before it stay created or updated, and\n" +
			"Emporix answers with the error of the one that failed (409 Conflict, exit 5, while a job\n" +
			"of it is running or when its metadata.version is stale).",
		example: "  epx search indexes import --file indexes.json --dry-run\n" +
			"  epx search indexes import --file indexes.json --wait --yes",
		req:      request{path: indexes.path + "/import"},
		scopes:   []string{searchManageScope},
		body:     &payload{},
		language: mapsOnly,
		answer:   job.answering(jobIDs{field: "jobId", list: true, of: []string{"type", "id"}}),
	}).command(a)
}
