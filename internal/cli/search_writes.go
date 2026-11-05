package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// perTypeWrites declares the writes of one item of a Search Service
// collection kept per type (perTypeCommand): update TYPE ID, an upsert of
// the JSON object in --file, whose localized text is sent as maps only
// (mapsOnly), and delete TYPE ID, both with searchManageScope.
// Each text field says what its write sends and what Emporix answers; what a
// type is (builtInTypes.sentence), what TYPE and ID are, how `epx api` sends
// it to an item kept under another spelling of a built-in type, how it is
// sent and the scope sentence follow.
type perTypeWrites struct {
	update  string   // what update sends
	example string   // a command line of update, indented by two spaces
	delete  string   // what delete sends
	answer  answerer // how update and delete end: by following the job that they start; by printing when nil
}

// commands returns update and delete of target, one item of a collection
// kept per type, whose TYPE takes types, which also say how to reach an item
// kept under another spelling of them.
func (p perTypeWrites) commands(a *app, target perTypeTarget, types exactTypes) []*cobra.Command {
	manage := []string{searchManageScope}
	typeAndArguments := "\n" + types.builtIn.sentence() + "\n" + target.arguments
	viaAPI := func(done, method, flags string) string {
		return "\n" + types.kept("A "+target.singular) + "is " + done + " with\n" +
			types.api(method, target.req, manage, flags) + "."
	}
	return []*cobra.Command{
		(&write{
			use: "update TYPE ID", short: "Upsert a " + target.singular, method: http.MethodPut,
			long:    p.update + typeAndArguments + ", or a new ID." + viaAPI("upserted", "put", " --file FILE"),
			example: p.example,
			req:     target.req, scopes: manage, body: &payload{}, language: mapsOnly, ids: target.ids,
			builtIn: types.builtIn, answer: p.answer,
		}).command(a),
		(&write{
			use: "delete TYPE ID", short: "Delete a " + target.singular, method: http.MethodDelete,
			long: p.delete + typeAndArguments + "." + viaAPI("deleted", "delete", ""),
			req:  target.req, scopes: manage, body: noBody{}, ids: target.ids, builtIn: types.builtIn, answer: p.answer,
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
		update: "Creates the search index ID of the type TYPE, or replaces it (an upsert), with the one\n" +
			"in --file, a JSON object with its fields (each with its path and how it is searched) and\n" +
			"optionally name and description. A job builds a new index, and one whose fields changed or\n" +
			"that is building or failed: Emporix then answers 202 Accepted with the job's id\n" +
			"(`epx search jobs get ID` reads the job), else 204 No Content. With metadata.version, as\n" +
			"`epx search indexes get TYPE ID -o json` prints it, Emporix replaces only that version;\n" +
			"another, or a job of the index that is still running, is a conflict (exit 5).",
		example: "  epx search indexes update VEHICLE vehicles --file index.json --dry-run\n" +
			"  epx search indexes update VEHICLE vehicles --file index.json --wait --yes",
		delete: "Starts the job that deletes the search index ID of the type TYPE: Emporix answers with\n" +
			"the job's id (`epx search jobs get ID` reads the job), and with 409 Conflict (exit 5) while\n" +
			"another job of the index is running.",
		answer: job.answering(jobIDs{field: "id"}),
	}
}

// savedSearchWrites declares the writes of one saved search.
func savedSearchWrites() perTypeWrites {
	return perTypeWrites{
		update: "Creates the saved search ID of the type TYPE, or replaces it (an upsert), with the one in\n" +
			"--file, a JSON object with index (the ID of a search index of TYPE), queries, filters or\n" +
			"both, and optionally name and description. Emporix answers a new one with its id (201\n" +
			"Created), and a replaced one with 204 No Content. Replacing one needs its metadata.version,\n" +
			"as `epx search queries get TYPE ID -o json` prints it; a stale one is a conflict (exit 5).\n" +
			"An unknown index is not found (exit 4).",
		example: "  epx search queries update VEHICLE by-make --file by-make.json --dry-run",
		delete:  "Deletes the saved search ID of the type TYPE; an unknown one is not found (exit 4).",
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
