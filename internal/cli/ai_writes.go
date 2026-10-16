package cli

import (
	"cmp"
	"net/http"
	"slices"

	"github.com/spf13/cobra"
)

// The writes of the AI Service's agents and of what they are built from:
// upserts, patches and deletes of one item (aiItemWrites), agents made from
// templates, the export and import of agents, and deletes of jobs. Text
// generation and chats with agents are in ai_chat.go.

// usedByAgents is the usage of --force of an update or patch of an item that
// enabled agents may use.
const usedByAgents = "disable it even where enabled agents use it, which disables them too"

// removeFromAgents is the usage of --force of a delete of an item that agents
// may use.
const removeFromAgents = "delete it even where agents use it"

// aiItemWrites declares the writes of one item of a collection of the AI
// Service, each with the scope that manages agents: update ID (PUT, an
// upsert), patch ID (PATCH) and delete ID (DELETE). The reference documents
// update and patch alike where they take force: they disable the item and
// what uses it.
type aiItemWrites struct {
	one      string       // the item with its article in summaries, e.g. "an MCP server"
	members  string       // what the JSON of update holds: sentences, each line ending with a line break
	language languageMode // whether its body holds localized text, as an agent's does, and how it is sent
	disable  string       // the usage of --force of update and patch (whom disabling it disables); none when empty
	remove   string       // the usage of --force of delete (whom deleting it unassigns it from)
}

// aiManagedGroup returns the group of r, a collection of the AI Service whose
// items w declares the writes of: list, get ID, search, update ID, patch ID,
// delete ID and the extra commands. Its summary names them unless r has its own.
func aiManagedGroup[T any](a *app, r resource[T], w aiItemWrites, extra ...*cobra.Command) *cobra.Command {
	r.short = cmp.Or(r.short, "Read, upsert, change and delete "+r.nouns())
	return newResourceCommand(a, r, slices.Concat([]*cobra.Command{r.querySearchCommand(a)},
		aiItemWriteCommands(a, r, w), extra)...)
}

// aiItemWriteCommands returns update ID, patch ID and delete ID of an item of
// r, as w declares them.
func aiItemWriteCommands[T any](a *app, r resource[T], w aiItemWrites) []*cobra.Command {
	item, scopes := request{path: r.itemPath()}, []string{aiManageScope}
	the := "the " + r.singular + " " + r.header[0]
	id := r.header[:1]
	source := r.header[0] + " is the " + r.header[0] + " that `epx " + r.listLine() + "` shows."
	read := "`epx " + r.commandLine(r.itemUse("get"), r.itemRequest()) + " -o json`"
	remove := itemDelete(r, "Delete "+w.one, "Deletes "+the+". Emporix refuses to delete one that others use, unless\n"+
		"--force unassigns it from them.\n"+source, scopes)
	remove.switches = forceSwitch(w.remove)
	return []*cobra.Command{
		(&write{
			use: r.itemUse("update"), short: "Upsert " + w.one, method: http.MethodPut,
			long: "Replaces " + the + " with the JSON in --file, or creates it with that ID where\n" +
				"none exists (an upsert): Emporix answers with the id of one it created and nothing for one it\n" +
				"replaced. With metadata.version, as " + read + " prints it,\n" +
				"Emporix replaces only that version: another is a conflict (exit 5).\n" + w.members + source,
			req: item, scopes: scopes, body: &payload{}, language: w.language, switches: forceSwitch(w.disable),
			ids: r.itemIDs(),
		}).command(a),
		(&write{
			use: r.itemUse("patch"), short: "Change members of " + w.one, method: http.MethodPatch,
			long: "Changes " + the + " with the JSON array of operations in --file, each with op\n" +
				"(ADD, REMOVE or REPLACE), path and value, e.g.\n" +
				`[{"op": "REPLACE", "path": "/enabled", "value": false}]; its other members keep their` + "\n" +
				"values.\n" + itemSentence(id, id, r.listLine()),
			req: item, scopes: scopes, body: operationsFile(), switches: forceSwitch(w.disable), ids: r.itemIDs(),
		}).command(a),
		remove.command(a),
	}
}

// forceSwitch returns --force with usage, which sends force=true, and
// force=false without it; none for an empty usage.
func forceSwitch(usage string) []paramFlag {
	if usage == "" {
		return nil
	}
	return []paramFlag{{param: forceParam, flag: forceParam, usage: usage}}
}

// agentWrites returns the writes of agents besides those of one agent: the
// attachment of a file to a chat, an agent made from a template, and the
// export and import of agents.
func agentWrites(a *app) []*cobra.Command {
	return []*cobra.Command{
		agentAttach().command(a),
		(&write{
			use:   "create TEMPLATE_ID",
			short: "Create an agent from a template",
			long: "Creates an agent from the agent template TEMPLATE_ID, the ID that `epx ai templates list`\n" +
				"shows, which must be enabled, with the JSON in --file: its name and description, its\n" +
				"userPrompt and, optionally, its id. Emporix answers with the id of the agent.",
			example: "  epx ai agents create complaint-template --file agent.json --dry-run",
			method:  http.MethodPost,
			req:     request{path: agenticPath + "templates/" + placeholderID + "/agents"},
			scopes:  []string{aiManageScope}, body: &payload{}, language: stringsOrMaps,
			ids: argumentIDs{"TEMPLATE_ID": aiTemplates().ids()},
		}).command(a),
		(&write{
			use:   "export",
			short: "Export agents, which creates an export job",
			long: "Exports the agents that --agent names (up to 100) with the tools and MCP servers they use,\n" +
				"and prints the export: data, the agents base64-encoded, and its checksum, which\n" +
				"`epx ai agents import` takes on another tenant. Emporix keeps each export as a job of type\n" +
				"EXPORT, the jobId of the response, which `epx ai jobs get JOB_ID` reads.",
			example: "  epx ai agents export --agent complaint-agent --yes > agents.json",
			method:  http.MethodPost,
			req:     request{path: agenticPath + "agents/export"},
			scopes:  []string{aiScope},
			body: &payload{typed: objectBody{
				listField("agentIds", "agent", "the `ID` of an agent to export (repeatable)").require(),
			}},
			readScoped: true,
		}).command(a),
		(&write{
			use:   importUse,
			short: "Import agents",
			long: "Imports the agents of an export, with the tools and MCP servers they use, from the JSON in\n" +
				"--file: the data and checksum that `epx ai agents export` prints, which Emporix checks.\n" +
				"Imported agents start disabled; Emporix answers with what it imported and the jobId of the\n" +
				"IMPORT job that it keeps (`epx ai jobs get JOB_ID`).",
			example: "  epx ai agents import --file agents.json --dry-run",
			method:  http.MethodPost,
			req:     request{path: agenticPath + "agents/import"},
			scopes:  []string{aiManageScope},
			body:    &payload{},
		}).command(a),
	}
}

// aiJobDelete deletes a job of jobs.
func aiJobDelete(jobs resource[aiJob]) *write {
	id := jobs.header[:1]
	return itemDelete(jobs, "Delete an agent job", "Deletes the agent job ID.\n"+itemSentence(id, id, jobs.listLine()),
		[]string{aiManageScope})
}
