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
	searchPerType     = searchService + "/search/" // the collections kept per type (perTypeCommand)
	searchScope       = "search.search_read"
	searchManageScope = "search.search_manage"
)

// searchDocumentTypes returns the types that Emporix defines whose
// documents the Search Service searches besides the instances of custom
// entity types, exactly as written, as Emporix keeps indexes and saved
// searches under a type as written (live: CUSTOMER addresses other indexes
// than customer), each with the read scope that finds all of its documents,
// which a search of them requires in place of the read scope of a custom
// entity type (documentsCommand).
func searchDocumentTypes() builtInTypes {
	return spelledTypes(builtInType{name: "customer", scopes: []string{customerReadScope}},
		builtInType{name: "legal-entity", scopes: []string{legalEntityReadScope}})
}

// searchTypes returns searchDocumentTypes without their scopes, as the
// commands of indexes and saved searches take them.
func searchTypes() builtInTypes { return searchDocumentTypes().withoutScopes() }

// badQuery says how the Search Service answers a malformed --query (live).
const badQuery = "Emporix answers a list's malformed --query with an internal error (500 Internal Server " +
	"Error, exit 1)."

// newSearchCommand reads the Search Service, which searches customers, legal
// entities and the instances of custom entity types: its indexes and saved
// searches, kept per type, and the jobs that create, update and delete the
// indexes.
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

// indexesCommand returns the group of search indexes, whose TYPE is one of
// types or a custom entity type, whose writes start the index jobs that job
// follows.
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

// indexJobs is the collection of the jobs that create, update and delete
// search indexes.
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

// documentsCommand returns `documents TYPE`, which reads the documents of a
// type, one of types or a custom entity type, that a saved search finds, or
// the search --body gives.
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
			"(exit 4). " + typeSentence(types) + " " +
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

// indexRef selects one search index to export: its type and its id.
type indexRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// indexSelection returns how one search index given as TYPE/ID is read as
// its indexRef. Its ID holds no slash; its type may. A type that rule, that
// of the TYPE argument of the commands of search indexes, does not accept is
// refused, as that argument refuses it (typeRule.accepts).
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

// perTypeCommand returns the group of a Search Service collection kept per
// type: list (of every type, at r.path), for-type TYPE, get TYPE ID and the
// writes of one item, update TYPE ID and delete TYPE ID, whose paths name the
// type before the collection, and extra commands. TYPE is one of types or
// one that list shows, and completes with those, ID one that for-type shows;
// the help of the group and of its reads ends r.long with what a type is
// (builtInTypes.sentence), as the writes add it to theirs, and with how to
// reach an item kept under another spelling of one of types (exactTypes).
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

// perTypeTarget is one item of a Search Service collection kept per type,
// which get, update and delete address: singular names it, req is its
// request, whose path names its type ({parent}) and ID, arguments says what
// TYPE and ID are, and ids lists them for completion.
type perTypeTarget struct {
	singular  string
	req       request
	arguments string
	ids       argumentIDs
}

// exactTypes are the built-in types that the TYPE of the collections kept per
// type of a service takes exactly as written, such as those of the Search
// Service, which keeps items under another spelling too, although TYPE
// refuses it; the help says how to reach such an item. otherSpelling is the
// first of builtIn in upper case, such a spelling (newExactTypes checks that
// TYPE refuses it).
type exactTypes struct {
	builtIn       builtInTypes
	otherSpelling string
}

// newExactTypes returns builtIn as the exactTypes of the collections of use.
// Where TYPE does not refuse the other spelling, as without built-in types,
// with types taken in any case or with a first one named in upper case,
// saying how to reach an item under it is a mistake in the declaration of
// use, which panics while the command tree is built, which every test of the
// tree does.
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

// kept returns, for help, subject, such as "A saved search", kept under
// another spelling of one of t, which its caller's sentence goes on from:
// "A saved search kept under another spelling of a built-in type, such as
// CUSTOMER,".
func (t exactTypes) kept(subject string) string {
	return subject + " kept under another spelling of a built-in type, such as " + t.otherSpelling + ","
}

// api returns, quoted, the `epx api` command line that sends method to the
// item of item, the request of one item of the collection, with another
// spelling of one of t as its type ({parent}) and the ID ID, with scopes and
// flags (none when empty).
func (t exactTypes) api(method string, item request, scopes []string, flags string) string {
	path := strings.NewReplacer(placeholderParent, t.otherSpelling, placeholderID, "ID").Replace(item.path)
	return "`epx api " + method + " " + path + " --scope " + strings.Join(scopes, " --scope ") + flags + "`"
}
