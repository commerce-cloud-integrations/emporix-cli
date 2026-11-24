package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

type perTypeWrites struct {
	update  string
	example string
	delete  string
	answer  answerer
}

func (p perTypeWrites) commands(a *app, target perTypeTarget, types exactTypes) []*cobra.Command {
	manage := []string{searchManageScope}
	typeAndArguments := types.builtIn.sentence() + " " + target.arguments
	viaAPI := func(done, method, flags string) string {
		return types.kept("A "+target.singular) + " is " + done + " with " +
			types.api(method, target.req, manage, flags) + "."
	}
	return []*cobra.Command{
		(&write{
			use: "update TYPE ID", short: "Upsert a " + target.singular, method: http.MethodPut,
			long:    joinSentences(p.update, typeAndArguments+", or a new ID.", viaAPI("upserted", "put", " --file FILE")),
			example: p.example,
			req:     target.req, scopes: manage, body: &payload{}, language: mapsOnly, ids: target.ids,
			builtIn: types.builtIn, answer: p.answer,
		}).command(a),
		(&write{
			use: "delete TYPE ID", short: "Delete a " + target.singular, method: http.MethodDelete,
			long: joinSentences(p.delete, typeAndArguments+".", viaAPI("deleted", "delete", "")),
			req:  target.req, scopes: manage, body: noBody{}, ids: target.ids, builtIn: types.builtIn, answer: p.answer,
		}).command(a),
	}
}

func indexJobWait(jobs resource[indexJob]) *jobWait {
	return followJob(jobs, (*indexJob).state, []string{"in_progress"}, "success")
}

func (j *indexJob) state() jobState {
	return jobState{ID: j.ID, Status: j.Status, Failure: j.Response}
}

func searchIndexWrites(job *jobWait) perTypeWrites {
	return perTypeWrites{
		update: "Creates the search index ID of the type TYPE, or replaces it (an upsert), with the one " +
			"in --file, a JSON object with its fields (each with its path and how it is searched) and " +
			"optionally name and description. A job builds a new index, and one whose fields changed or " +
			"that is building or failed: Emporix then answers 202 Accepted with the job's id " +
			"(`epx search jobs get ID` reads the job), else 204 No Content. With metadata.version, as " +
			"`epx search indexes get TYPE ID -o json` prints it, Emporix replaces only that version; " +
			"another, or a job of the index that is still running, is a conflict (exit 5).",
		example: "  epx search indexes update VEHICLE vehicles --file index.json --dry-run\n" +
			"  epx search indexes update VEHICLE vehicles --file index.json --wait --yes",
		delete: "Starts the job that deletes the search index ID of the type TYPE: Emporix answers with " +
			"the job's id (`epx search jobs get ID` reads the job), and with 409 Conflict (exit 5) while " +
			"another job of the index is running. An unknown one is not found (exit 4).",
		answer: job.answering(jobIDs{field: "id"}),
	}
}

func savedSearchWrites() perTypeWrites {
	return perTypeWrites{
		update: "Creates the saved search ID of the type TYPE, or replaces it (an upsert), with the one in " +
			"--file, a JSON object with index (the ID of a search index of TYPE), queries, filters or " +
			"both, and optionally name and description. Emporix answers a new one with its id (201 " +
			"Created), and a replaced one with 204 No Content. Replacing one needs its metadata.version, " +
			"as `epx search queries get TYPE ID -o json` prints it; a stale one is a conflict (exit 5), and " +
			"without it Emporix refuses an existing ID (400 Bad Request, exit 1). A read prints the index " +
			"as indexId; the body names it as index. An unknown index is not found (exit 4).",
		example: "  epx search queries update VEHICLE by-make --file by-make.json --dry-run",
		delete:  "Deletes the saved search ID of the type TYPE; an unknown one is not found (exit 4).",
	}
}

func searchIndexImport(a *app, indexes resource[searchIndex], job *jobWait) *cobra.Command {
	return (&write{
		use: importUse, short: "Import search indexes", method: http.MethodPost,
		long: "Creates or updates each search index of the export package in --file, as update does. The " +
			"package is the whole response of `epx search indexes export -o json`, with its exportedAt " +
			"(Emporix answers 400 without it). Emporix answers with a " +
			"result per index, in the order of the package, with the id of the job that builds it " +
			"where one started. When an index fails, those before it stay created or updated, and " +
			"Emporix answers with the error of the one that failed (409 Conflict, exit 5, while a job " +
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
