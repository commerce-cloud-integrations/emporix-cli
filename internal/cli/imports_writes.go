package cli

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// The writes of the Import Service start, retry and cancel import runs and
// set or delete the schedule of a configuration. Emporix allows them with
// importScope, the scope that its reads need too.

// activeRunConflict is what a 409 Conflict of a write that starts a run
// means: the configuration has a run that has not ended.
const activeRunConflict = "the configuration has a run QUEUED, RUNNING or RESUMABLE already: wait until it ends, " +
	"or cancel it"

// importRunJob returns how --wait follows an import run that a write starts:
// it reads the run as `imports runs get` does, in the run's details, until it
// is no longer QUEUED, RUNNING or RESUMABLE (interrupted; it resumes by
// itself), and succeeds only when it ended SUCCEEDED. Emporix answers the
// write with the run itself (oneItem).
func importRunJob(runs resource[importRun]) *jobWait {
	return followJob(runs, (*importRun).state, []string{"QUEUED", "RUNNING", "RESUMABLE"}, "SUCCEEDED").
		answering(heldJob(jobDecoder(runs, oneItem, (*importRun).state)))
}

// state returns what --wait follows of r: how many records it read, where
// it says, and its message, why it failed.
func (r *importRun) state() jobState {
	s := jobState{ID: r.ID, Status: r.Status, Failure: r.Message}
	if r.RecordsRead != "" {
		s.Progress = r.RecordsRead.String() + " read"
	}
	return s
}

// importConfigurationWrites returns the writes of an import configuration of
// configurations: trigger, which starts a run that job follows, and
// update-schedule and delete-schedule.
func importConfigurationWrites(a *app, configurations resource[importConfig], job *jobWait) []*cobra.Command {
	scopes := importScopes()
	configuration := configurations.argument()
	of := configurations.itemPath()
	source := "\n" + configurations.argumentSource() + "."
	return []*cobra.Command{
		importTrigger(configurations, job).command(a),
		(&write{
			use: "update-schedule " + configuration, short: "Upsert the run schedule of a configuration",
			method: http.MethodPut,
			long: "Sets when the configuration " + configuration + " runs by itself, creating its schedule where\n" +
				"it has none (an upsert), and prints the schedule. --cron is a Spring cron expression with\n" +
				"six fields, seconds first, e.g. \"0 0 2 * * *\" for 02:00 every day (Emporix refuses one with\n" +
				"five: 400 Bad Request), evaluated in --timezone, an IANA time zone such as Europe/Berlin.\n" +
				"The schedule runs the configuration unless --enabled=false, which keeps it without running\n" +
				"it: epx always sends enabled, as the schedule is replaced whole. --file gives the body\n" +
				"instead." + source,
			example: "  epx imports configurations update-schedule 7c9e6679-7425-40de-944b-e07fc1f90ae7 " +
				"--cron \"0 0 2 * * *\" --timezone Europe/Berlin",
			req: request{path: of + "/schedule"}, scopes: scopes,
			body: &payload{typed: objectBody{
				stringField("cron", "cron", "the Spring cron `EXPRESSION` of the schedule, six fields").require(),
				stringField("timezone", "timezone", "the IANA time `ZONE` the expression is evaluated in"),
				onField("enabled", "enabled", "whether the schedule runs the configuration (--enabled=false: not)"),
			}},
		}).command(a),
		(&write{
			use: "delete-schedule " + configuration, short: "Delete the run schedule of a configuration",
			method: http.MethodDelete,
			long: "Deletes the schedule of the configuration " + configuration + ", which then runs only when it\n" +
				"is triggered. Emporix answers alike when there is no schedule, also of a configuration that\n" +
				"was deleted.",
			req: request{path: of + "/schedule"}, scopes: scopes, body: noBody{},
		}).command(a),
	}
}

// importTrigger starts a run of an import configuration of configurations,
// which job follows.
func importTrigger(configurations resource[importConfig], job *jobWait) *write {
	configuration := configurations.argument()
	return &write{
		use: "trigger " + configuration, short: "Start an import run of a configuration", method: http.MethodPost,
		long: "Starts an import run of the configuration " + configuration + ", a DELTA run unless --mode FULL,\n" +
			"and prints the run, QUEUED until a worker of the Import Service takes it. --file gives the\n" +
			"body instead, e.g. {\"mode\": \"FULL\", \"dryRun\": true, \"streamIds\": [\"ID\"]} for a dry run of\n" +
			"some streams, which maps and checks records but writes none. A configuration that has a run\n" +
			"QUEUED, RUNNING or RESUMABLE already is a conflict (exit 5). Emporix answers a run that it\n" +
			"refuses, e.g. of streams whose mappings were never published, as any other: the run ends\n" +
			"ABORTED, which --wait reports.\n" + configurations.argumentSource() + ".",
		example: "  epx imports configurations trigger 7c9e6679-7425-40de-944b-e07fc1f90ae7 --mode FULL --wait",
		req:     request{path: configurations.itemPath() + "/runs"}, scopes: importScopes(), answer: job,
		conflict: activeRunConflict,
		body: &payload{typed: objectBody{choiceField("mode", "mode",
			"the `MODE` of the run, DELTA when not given", "FULL", "DELTA")}},
	}
}

// importRunWrites returns the writes of an import run of runs: cancel, and
// retry, which starts a run that job follows.
func importRunWrites(a *app, runs resource[importRun], job *jobWait) []*cobra.Command {
	return []*cobra.Command{
		(&write{
			use: runs.itemUse("cancel"), short: "Cancel an import run", method: http.MethodPost,
			long: "Asks the import run ID to stop, and prints what Emporix answers: whether it accepted the\n" +
				"cancellation. A QUEUED run is CANCELLED at once; a running one stops when it can, or with\n" +
				"--force at once, which also unblocks its configuration. A RESUMABLE run stops only with\n" +
				"--force: without it, a conflict (exit 5). Emporix answers a run that is unknown or ended\n" +
				"already with 409 Conflict and \"accepted\": false, which epx prints as the answer (exit 5)." +
				runSource(runs),
			req: request{path: runs.itemPath() + "/cancel"}, scopes: importScopes(), body: noBody{},
			switches: forceSwitch("stop the run at once"), answers: emporix.Answers{FailureAnswered: true},
			conflict: cancelConflict,
		}).command(a),
		importRetry(runs, job).command(a),
	}
}

// cancelConflict is what a 409 Conflict of a cancel means: a RESUMABLE run
// stops only with --force, and an unknown or ended run is answered with
// accepted false, which Emporix gives the same 409 for.
const cancelConflict = "a RESUMABLE run stops only with --force; an unknown or ended run is answered " +
	"with \"accepted\": false"

// importRetry runs the failed records of an import run of runs again, in a
// run that job follows.
func importRetry(runs resource[importRun], job *jobWait) *write {
	return &write{
		use: runs.itemUse("retry"), short: "Run the failed records of an import run again", method: http.MethodPost,
		long: "Starts a run that imports again only the records that failed in the import run ID, and\n" +
			"prints it, QUEUED until a worker takes it; the run ID keeps its counters. A configuration\n" +
			"that has a run QUEUED, RUNNING or RESUMABLE already is a conflict (exit 5)." + runSource(runs),
		req: request{path: runs.itemPath() + "/retry"}, scopes: importScopes(), body: noBody{}, answer: job,
		conflict: activeRunConflict,
	}
}

// runSource says, in the help of a write of an import run of runs, where its
// ID comes from.
func runSource(runs resource[importRun]) string {
	return "\nID is the one that `epx " + runs.listLine() + "` shows."
}
