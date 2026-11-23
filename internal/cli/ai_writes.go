package cli

import (
	"cmp"
	"net/http"
	"slices"

	"github.com/spf13/cobra"
)

const usedByAgents = "disable it even where enabled agents use it, which disables them too"

const removeFromAgents = "delete it even where agents use it"

// aiItemWrites declares the writes of one item of a collection of the AI
// Service, each with the scope that manages agents: update ID (PUT, an
// upsert), patch ID (PATCH) and delete ID (DELETE). The reference documents
// update and patch alike where they take force: they disable the item and
// what uses it.
type aiItemWrites struct {
	summaryNoun  string
	members      string
	language     languageMode
	disableUsage string
	deleteUsage  string
}

func aiManagedGroup[T any](a *app, r resource[T], w aiItemWrites, extra ...*cobra.Command) *cobra.Command {
	r.short = cmp.Or(r.short, "Read, upsert, change and delete "+r.nouns())
	return newResourceCommand(a, r, slices.Concat([]*cobra.Command{r.querySearchCommand(a)},
		aiItemWriteCommands(a, r, w), extra)...)
}

func aiItemWriteCommands[T any](a *app, r resource[T], w aiItemWrites) []*cobra.Command {
	item, scopes := request{path: r.itemPath()}, []string{aiManageScope}
	the := "the " + r.singular + " " + r.header[0]
	id := r.header[:1]
	source := r.header[0] + " is the " + r.header[0] + " that `epx " + r.listLine() + "` shows."
	read := "`epx " + r.commandLine(r.itemUse("get"), r.itemRequest()) + " -o json`"
	remove := itemDelete(r, "Delete "+w.summaryNoun, "Deletes "+the+". Emporix refuses to delete one that others "+
		"use, unless --force unassigns it from them. "+source+" "+unknownDeletedHelp, scopes)
	remove.switches = forceSwitch(w.deleteUsage)
	return []*cobra.Command{
		(&write{
			use: r.itemUse("update"), short: "Upsert " + w.summaryNoun, method: http.MethodPut,
			long: joinSentences("Replaces "+the+" with the JSON in --file, or creates it with that ID where "+
				"none exists (an upsert): Emporix answers with the id of one it created and nothing for one it "+
				"replaced. With metadata.version, as "+read+" prints it, "+
				"Emporix replaces only that version: another is a conflict (exit 5).", w.members, source),
			req: item, scopes: scopes, body: &payload{}, language: w.language, switches: forceSwitch(w.disableUsage),
			ids: r.itemIDs(),
		}).command(a),
		(&write{
			use: r.itemUse("patch"), short: "Change members of " + w.summaryNoun, method: http.MethodPatch,
			long: "Changes " + the + " with the JSON array of operations in --file, each with op " +
				"(ADD, REMOVE or REPLACE), path and value, e.g. " +
				literal(`[{"op": "REPLACE", "path": "/enabled", "value": false}]`) + "; its other members keep their values. " +
				itemSentence(id, id, r.listLine()),
			req: item, scopes: scopes, body: operationsFile(), switches: forceSwitch(w.disableUsage), ids: r.itemIDs(),
		}).command(a),
		remove.command(a),
	}
}

func forceSwitch(usage string) []paramFlag {
	if usage == "" {
		return nil
	}
	return []paramFlag{{param: forceParam, flag: forceParam, usage: usage}}
}

func agentWrites(a *app) []*cobra.Command {
	return []*cobra.Command{
		agentAttach().command(a),
		(&write{
			use:   "create TEMPLATE_ID",
			short: "Create an agent from a template",
			long: "Creates an agent from the agent template TEMPLATE_ID, the ID that `epx ai templates list` " +
				"shows, which must be enabled, with the JSON in --file: its name and description, its " +
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
			long: "Exports the agents that --agent names (up to 100) with the tools and MCP servers they use, " +
				"and prints the export: data, the agents base64-encoded, and its checksum, which " +
				"`epx ai agents import` takes on another tenant. Emporix keeps each export as a job of type " +
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
			long: "Imports the agents of an export, with the tools and MCP servers they use, from the JSON in " +
				"--file: the data and checksum that `epx ai agents export` prints, which Emporix checks. " +
				"Imported agents start disabled; Emporix answers with what it imported and the jobId of the " +
				"IMPORT job that it keeps (`epx ai jobs get JOB_ID`).",
			example: "  epx ai agents import --file agents.json --dry-run",
			method:  http.MethodPost,
			req:     request{path: agenticPath + "agents/import"},
			scopes:  []string{aiManageScope},
			body:    &payload{},
		}).command(a),
	}
}

func aiJobDelete(jobs resource[aiJob]) *write {
	return itemDelete(jobs, "Delete an agent job", "Deletes the agent job ID. ID is the ID that `epx "+
		jobs.listLine()+"` shows. "+unknownDeletedHelp, []string{aiManageScope})
}
