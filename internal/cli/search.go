package cli

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type searchIndex struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Name     localized    `json:"name"`
	Status   string       `json:"status"`
	Fields   []indexField `json:"fields"`
	Metadata itemMetadata `json:"metadata"`
}

type indexField struct {
	Path string `json:"path"`
}

func fieldPaths(fields []indexField) string {
	return joinEach(fields, ", ", func(f indexField) string { return f.Path })
}

type savedSearch struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Name     localized    `json:"name"`
	IndexID  string       `json:"indexId"`
	Metadata itemMetadata `json:"metadata"`
}

type indexJob struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Status   string       `json:"status"`
	IndexID  string       `json:"indexId"`
	Response string       `json:"response"`
	Metadata itemMetadata `json:"metadata"`
}

type searchDocument struct {
	ID    string      `json:"id"`
	Name  localized   `json:"name"`
	Score json.Number `json:"_score"`
}

const (
	searchUse         = "search"
	searchService     = "/search/{tenant}"
	searchPerType     = searchService + "/search/"
	searchScope       = "search.search_read"
	searchManageScope = "search.search_manage"
)

func searchDocumentTypes() builtInTypes {
	return spelledTypes(builtInType{name: "customer", scopes: []string{customerReadScope}},
		builtInType{name: "legal-entity", scopes: []string{legalEntityReadScope}})
}

func searchTypes() builtInTypes { return searchDocumentTypes().withoutScopes() }

const badQuery = "Emporix answers a list's malformed --query with an internal error (500 Internal Server " +
	"Error, exit 1)."

func newSearchCommand(a *app) *cobra.Command {
	types := newExactTypes(searchTypes(), searchUse)
	queries := resource[savedSearch]{
		use:   "queries",
		noun:  "saved searches",
		short: "Read, upsert and delete saved searches",
		under: searchUse,
		long: "Reads saved searches: queries and filters (only in `-o json`) stored to run against a search " +
			"INDEX of a type. " + badQuery,
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
		"Reads the Search Service (a preview at Emporix), which searches customers, legal entities "+
			"and the instances of custom entity types: the documents a search finds, its indexes and "+
			"saved searches, kept per type, and the jobs that create, update and delete the indexes; "+
			"changes the indexes and saved searches. "+types.builtIn.sentence(),
		documentsCommand(a, searchDocumentTypes()),
		indexesCommand(a, indexJobWait(jobs), types),
		perTypeCommand(a, queries, savedSearchWrites(), types),
		newResourceCommand(a, jobs))
}

func indexesCommand(a *app, job *jobWait, types exactTypes) *cobra.Command {
	indexes := resource[searchIndex]{
		use:   "indexes",
		noun:  "search indexes",
		short: "Read, upsert, delete, export and import search indexes",
		under: searchUse,
		long: "Reads search indexes: which fields of the documents of a type can be searched. " +
			"STATUS is building, ready or failed and FIELDS the indexed fields (`-o json` shows how each is " +
			"searched). --query matches STATUS in upper case (`status:READY`; `status:ready` matches no index). " + badQuery,
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
		long: "Reads an export package of the search indexes that --index names, as TYPE/ID (the TYPE and " +
			"ID that `epx search indexes list` shows): their configuration, without status or metadata. " +
			"Emporix answers an unknown index as not found (exit 4). " + types.builtIn.sentence() + " " +
			types.kept("A search index") + " is exported with --body.",
		scopes: []string{searchManageScope},
		selection: objectsValue("index", "a search index to export, as `TYPE/ID` (repeatable)",
			indexSelection(newTypeRule(indexes.scopes, types.builtIn))),
	})
	return perTypeCommand(a, indexes, searchIndexWrites(job), types, indexExport,
		searchIndexImport(a, indexes, job))
}

func indexJobs() resource[indexJob] {
	return resource[indexJob]{
		use:   "jobs",
		noun:  "index jobs",
		under: searchUse,
		long: "Reads the jobs that create, update and delete search indexes: TYPE is create_index, " +
			"update_index or delete_index, STATUS in_progress, success or failure as the table and JSON " +
			"show them, and RESULT what the job recorded when it finished. --query matches the upper-case " +
			"form (`status:SUCCESS` and `type:UPDATE_INDEX` are checked; a lower-case filter, such as " +
			"`status:success`, matches no job). " + badQuery,
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

func documentsCommand(a *app, types builtInTypes) *cobra.Command {
	documents := reader[searchDocument]{
		scopes: []string{searchScope, customTypeReadScope}, // the reference also requires one that reads the documents
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
		short: "Search customers, legal entities or the instances of a custom entity type",
		long: "Reads the documents of a type that the saved search --saved-search finds with the text of " +
			"--text (`epx search queries for-type TYPE` lists them): customers, legal entities or the " +
			"instances of a custom entity type, as a search index holds them. SCORE is how well a document " +
			"matches; its other fields are only in `-o json`. --body gives a search of an index by its own " +
			"queries and filters instead. An unknown saved search, or index of --body, is not found " +
			"(exit 4). --sort sorts only by fields that the index holds with exact true (Emporix leaves " +
			"the others unsorted and refuses id, 400 Bad Request). Emporix orders documents of the same " +
			"SCORE differently on each request, so a search that finds more than one page of documents " +
			"(100) can repeat one on a later page: epx then fails with exit 1 (with --stream, after " +
			"printing the pages before it). " + typeSentence(types) + " " +
			"Known limitation: for customer or legal-entity, epx requests the read scope that " +
			"finds every document, so a credential that holds only a restricted one, such as " +
			"customer.customer_read--DE, or only the own-read scope, such as customer.customer_read_own, " +
			"fails with exit code 3.",
		req:     request{path: searchPerType + placeholderID},
		ids:     argumentIDs{columnType: customEntities().ids()},
		builtIn: types,
		fetch: listing{filters: listFilters{kinds: sortable}}.search(
			stringField("searchQueryId", "saved-search", "the `ID` of the saved search to run").require(),
			stringField("query", "text", "the `text` to search for in every term of the saved search").require(),
		),
	})
}

type indexRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func indexSelection(rule typeRule) func(value string) (indexRef, error) {
	return func(value string) (indexRef, error) {
		slash := strings.LastIndex(value, "/")
		if slash <= 0 || slash == len(value)-1 {
			return indexRef{}, errors.New("use TYPE/ID")
		}
		if err := rule.accepts(value[:slash]); err != nil {
			return indexRef{}, err
		}
		return indexRef{Type: value[:slash], ID: value[slash+1:]}, nil
	}
}

func perTypeCommand[T any](a *app, r resource[T], writes perTypeWrites, types exactTypes,
	extra ...*cobra.Command) *cobra.Command {
	ofType := request{path: searchPerType + placeholderID + "/" + r.use}
	rd, listed, listedTypes := r.reader(), r.listing(), r.idsIn(columnType)
	target := perTypeTarget{
		singular: r.singular,
		req:      request{path: searchPerType + placeholderParent + "/" + r.use + "/" + placeholderID},
		arguments: "TYPE and ID are a " + r.singular + "'s type and ID, as `epx " + r.under + " " + r.use +
			" list` shows them",
		ids: argumentIDs{columnType: listedTypes, columnID: listedIDs(rd, ofType, listed, columnID)},
	}
	r.long = joinSentences(r.long, types.builtIn.sentence(), types.kept("A "+r.singular)+" is listed too; `epx api` "+
		"reaches it by its path, such as "+types.api("get", target.req, r.scopes, "")+".")
	return newGroup(r.use, r.summary(), r.long, slices.Concat([]*cobra.Command{
		r.listCommand(a),
		rd.command(a, &command{
			use:   "for-type TYPE",
			short: "List the " + r.nouns() + " of a type",
			long: r.long + " TYPE is the type of " + r.nouns() + ", as `epx " + r.listLine() +
				"` shows it; an unknown type lists nothing.",
			req:     ofType,
			fetch:   listed,
			ids:     argumentIDs{columnType: listedTypes},
			builtIn: types.builtIn,
		}),
		rd.command(a, &command{
			use:     "get TYPE ID",
			short:   "Show one " + r.singular,
			long:    joinSentences(r.long, target.arguments+"; an unknown one is not found (exit 4)."),
			req:     target.req,
			fetch:   single{items: oneItem},
			ids:     target.ids,
			builtIn: types.builtIn,
		}),
	}, writes.commands(a, target, types), extra)...)
}

type perTypeTarget struct {
	singular  string
	req       request
	arguments string
	ids       argumentIDs
}

type exactTypes struct {
	builtIn       builtInTypes
	otherSpelling string
}

func newExactTypes(builtIn builtInTypes, use string) exactTypes {
	other := ""
	if len(builtIn.types) > 0 {
		other = strings.ToUpper(builtIn.types[0].name)
	}
	if builtIn.spelling(other) == nil {
		panic(use + ": says how to reach an item under another spelling of a built-in type, but TYPE does not " +
			"refuse the upper-case spelling of its first built-in type, if it has one")
	}
	return exactTypes{builtIn: builtIn, otherSpelling: other}
}

func (t exactTypes) kept(subject string) string {
	return subject + " kept under another spelling of a built-in type, such as " + t.otherSpelling + ","
}

func (t exactTypes) api(method string, item request, scopes []string, flags string) string {
	path := strings.NewReplacer(placeholderParent, t.otherSpelling, placeholderID, "ID").Replace(item.path)
	return "`epx api " + method + " " + path + " --scope " + strings.Join(scopes, " --scope ") + flags + "`"
}
