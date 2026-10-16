package cli

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// searchIndex is the part of a Search Service index shown in tables.
type searchIndex struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Name     localized    `json:"name"`
	Status   string       `json:"status"`
	Fields   []indexField `json:"fields"`
	Metadata itemMetadata `json:"metadata"`
}

// indexField is an indexed field; tables show its path, -o json also how it
// is searched.
type indexField struct {
	Path string `json:"path"`
}

// fieldPaths returns the paths of indexed fields (a long value, which
// tables cut).
func fieldPaths(fields []indexField) string {
	return joinEach(fields, ", ", func(f indexField) string { return f.Path })
}

// savedSearch is the part of a saved search shown in tables; its queries and
// filters are only in -o json.
type savedSearch struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Name     localized    `json:"name"`
	IndexID  string       `json:"indexId"`
	Metadata itemMetadata `json:"metadata"`
}

// indexJob is the part of a Search Service index job shown in tables.
type indexJob struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Status   string       `json:"status"`
	IndexID  string       `json:"indexId"`
	Response string       `json:"response"`
	Metadata itemMetadata `json:"metadata"`
}

// searchDocument is the part of a document that a search finds shown in
// tables; its other fields are only in -o json.
type searchDocument struct {
	ID    string      `json:"id"`
	Name  localized   `json:"name"`
	Score json.Number `json:"_score"` // kept as sent
}

// Name, paths and scopes of the Search Service: searchScope of its reads,
// searchManageScope of its writes and of the export of search indexes, a
// read that Emporix allows only with it.
const (
	searchUse         = "search"
	searchService     = "/search/{tenant}"
	searchPerType     = searchService + "/search/" // the collections kept per custom entity type
	searchScope       = "search.search_read"
	searchManageScope = "search.search_manage"
)

// badQuery says how the Search Service answers a malformed --query (live).
const badQuery = "\nEmporix answers a list's malformed --query with an internal error (500 Internal Server\n" +
	"Error, exit 1)."

// newSearchCommand reads the Search Service, which searches the instances of
// custom entity types: its indexes and saved searches, kept per type, and
// the jobs that create, update and delete the indexes.
func newSearchCommand(a *app) *cobra.Command {
	queries := resource[savedSearch]{
		use:   "queries",
		noun:  "saved searches",
		short: "Read, upsert and delete saved searches",
		under: searchUse,
		long: "Reads saved searches: queries and filters (only in -o json) stored to run against the\n" +
			"search INDEX of a custom entity type." + badQuery,
		path:    searchPerType + "queries",
		scopes:  []string{searchScope},
		filters: listFilters{kinds: filterable},
		view: view[savedSearch]{
			singular:  "saved search",
			header:    []string{columnID, columnType, columnName, columnIndex, columnVersion},
			localized: true,
			row: func(s savedSearch, lang language) []string {
				return []string{s.ID, s.Type, s.Name.in(lang), s.IndexID, s.Metadata.Version.String()}
			},
		},
	}
	jobs := indexJobs()
	return newGroup(searchUse, "Search documents; read and change search indexes and saved searches",
		"Reads the Search Service, which searches the instances of custom entity types (a preview\n"+
			"at Emporix): the documents a search finds, its indexes and saved searches, kept per type\n"+
			"(`epx custom-entities list` shows the types), and the jobs that create, update and delete\n"+
			"the indexes; changes the indexes and saved searches.",
		documentsCommand(a),
		indexesCommand(a, indexJobWait(jobs)),
		perTypeCommand(a, queries, savedSearchWrites()),
		newResourceCommand(a, jobs))
}

// indexesCommand returns the group of search indexes, whose writes start the
// index jobs that job follows.
func indexesCommand(a *app, job *jobWait) *cobra.Command {
	indexes := resource[searchIndex]{
		use:   "indexes",
		noun:  "search indexes",
		short: "Read, upsert, delete, export and import search indexes",
		under: searchUse,
		long: "Reads search indexes: which fields of a custom entity type's instances can be searched.\n" +
			"STATUS is building, ready or failed and FIELDS the indexed fields (-o json shows how each is\n" +
			"searched). --query matches STATUS in upper case (`status:READY`; `status:ready` matches no index)." + badQuery,
		path:    searchPerType + "indexes",
		scopes:  []string{searchScope},
		filters: listFilters{kinds: filterable},
		view: view[searchIndex]{
			singular:  "search index",
			header:    []string{columnID, columnType, columnName, columnStatus, columnFields, columnVersion},
			long:      []string{columnFields},
			localized: true,
			row: func(i searchIndex, lang language) []string {
				return []string{i.ID, i.Type, i.Name.in(lang), i.Status, fieldPaths(i.Fields), i.Metadata.Version.String()}
			},
		},
	}
	indexExport := exportCommand(a, indexes.path, &export{
		short: "Export search indexes",
		long: "Reads an export package of the search indexes that --index names, as TYPE/ID (the TYPE and\n" +
			"ID that `epx search indexes list` shows): their configuration, without status or metadata.\n" +
			"Emporix answers an unknown index as not found (exit 4).",
		scopes: []string{searchManageScope},
		selection: objectsValue("index", "a search index to export, as `TYPE/ID` (repeatable)",
			indexSelection),
	})
	return perTypeCommand(a, indexes, searchIndexWrites(job), indexExport, searchIndexImport(a, indexes, job))
}

// indexJobs is the collection of the jobs that create, update and delete
// search indexes.
func indexJobs() resource[indexJob] {
	return resource[indexJob]{
		use:   "jobs",
		noun:  "index jobs",
		under: searchUse,
		long: "Reads the jobs that create, update and delete search indexes: TYPE is create_index,\n" +
			"update_index or delete_index, STATUS in_progress, success or failure as the table and JSON\n" +
			"show them, and RESULT what the job recorded when it finished. --query matches the upper-case\n" +
			"form (`status:SUCCESS` and `type:UPDATE_INDEX` are checked; a lower-case filter, such as\n" +
			"`status:success`, matches no job)." + badQuery,
		path:    searchService + "/jobs",
		scopes:  []string{searchScope},
		filters: listFilters{kinds: filterable},
		view: view[indexJob]{
			singular: "index job",
			header:   []string{columnID, columnType, columnStatus, columnIndex, columnCreated, columnModified, "RESULT"},
			long:     []string{"RESULT"},
			row: func(j indexJob, _ language) []string {
				return []string{
					j.ID, j.Type, j.Status, j.IndexID, j.Metadata.CreatedAt, j.Metadata.ModifiedAt,
					j.Response,
				}
			},
		},
	}
}

// documentsCommand returns `documents TYPE`, which reads the documents of a
// custom entity type that a saved search finds, or the search --body gives.
func documentsCommand(a *app) *cobra.Command {
	documents := reader[searchDocument]{
		scopes: []string{searchScope, customTypeReadScope}, // the reference also requires one that reads the instances
		view: view[searchDocument]{
			singular:  "search document",
			header:    []string{columnID, columnName, "SCORE"},
			localized: true,
			row: func(d searchDocument, lang language) []string {
				return []string{d.ID, d.Name.in(lang), d.Score.String()}
			},
		},
	}
	return documents.command(a, &command{
		use:   "documents TYPE",
		short: "Search the documents of a custom entity type",
		long: "Reads the documents of a custom entity type, its instances as a search index holds them,\n" +
			"that the saved search --saved-search finds with the text of --text (`epx search queries\n" +
			"for-type TYPE` lists them). SCORE is how well a document matches; its other fields are\n" +
			"only in -o json. --body gives a search of an index by its own queries and filters\n" +
			"instead.\n" + typeSentence("TYPE"),
		req: request{path: searchPerType + placeholderID},
		ids: argumentIDs{columnType: customEntities().ids()},
		fetch: listing{filters: listFilters{kinds: sortable}}.search(
			stringField("searchQueryId", "saved-search", "the `ID` of the saved search to run").require(),
			stringField("query", "text", "the `text` to search for in every term of the saved search").require(),
		),
	})
}

// indexRef selects one search index to export: its custom entity type and
// its id.
type indexRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// indexSelection returns the indexRef of one search index given as TYPE/ID;
// its id holds no slash, its type may.
func indexSelection(value string) (indexRef, error) {
	slash := strings.LastIndex(value, "/")
	if slash <= 0 || slash == len(value)-1 {
		return indexRef{}, errors.New("use TYPE/ID")
	}
	return indexRef{Type: value[:slash], ID: value[slash+1:]}, nil
}

// perTypeCommand returns the group of a Search Service collection kept per
// custom entity type: list (of every type, at r.path), for-type TYPE, get
// TYPE ID and the writes of one item, update TYPE ID and delete TYPE ID,
// whose paths name the type before the collection, and extra commands. TYPE
// is one that list shows, and completes with those, ID one that for-type
// shows.
func perTypeCommand[T any](a *app, r resource[T], writes perTypeWrites, extra ...*cobra.Command) *cobra.Command {
	ofType := request{path: searchPerType + placeholderID + "/" + r.use}
	item := request{path: searchPerType + placeholderParent + "/" + r.use + "/" + placeholderID}
	arguments := "TYPE and ID are a " + r.singular + "'s custom entity type and ID, as `epx " + r.under + " " +
		r.use + " list`\nshows them"
	rd, listed, types := r.reader(), r.listing(), r.idsIn(columnType)
	ids := argumentIDs{columnType: types, columnID: listedIDs(rd, ofType, listed, columnID)}
	return newGroup(r.use, r.summary(), r.long, slices.Concat([]*cobra.Command{
		r.listCommand(a),
		rd.command(a, &command{
			use:   "for-type TYPE",
			short: "List the " + r.nouns() + " of a custom entity type",
			long: r.long + "\nTYPE is the custom entity type of " + r.nouns() + ", as `epx " + r.listLine() +
				"` shows it; an\nunknown type lists nothing.",
			req:   ofType,
			fetch: listed,
			ids:   argumentIDs{columnType: types},
		}),
		rd.command(a, &command{
			use:   "get TYPE ID",
			short: "Show one " + r.singular,
			long:  r.long + "\n" + arguments + "; an unknown one is not found (exit 4).",
			req:   item,
			fetch: single{items: oneItem},
			ids:   ids,
		}),
	}, writes.commands(a, r.singular, item, arguments, ids), extra)...)
}
