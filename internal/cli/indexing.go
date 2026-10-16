package cli

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// indexProvider is the part of an Indexing Service provider configuration
// shown in tables. It deliberately has no key fields: tables leave the search
// and write keys out, and -o json prints them.
type indexProvider struct {
	Provider      string `json:"provider"`
	Active        *bool  `json:"active"`
	ApplicationID string `json:"applicationId"`
	IndexName     string `json:"indexName"`
}

// reindexJob is the part of an Indexing Service reindex job shown in tables.
type reindexJob struct {
	ID             string       `json:"id"`
	EntityType     string       `json:"entityType"`
	Status         string       `json:"status"`
	ProcessedCount json.Number  `json:"processedCount"`
	ExpectedCount  json.Number  `json:"expectedCount"`
	Message        string       `json:"message"`
	Metadata       itemMetadata `json:"metadata"`
}

// indexingUse is the name of the group of the Indexing Service.
const indexingUse = "indexing"

// The scopes of the Indexing Service: of the tenant's own reads, and of its
// writes, which change or delete a provider or start a reindex.
const (
	indexingScope       = "indexing.search_view"
	indexingManageScope = "indexing.search_manage"
)

// providerView is the table of index providers; singular names them.
func providerView(singular string) view[indexProvider] {
	return view[indexProvider]{
		singular: singular,
		header:   []string{columnProvider, columnActive, "APPLICATION ID", columnIndex},
		row: func(p indexProvider, _ language) []string {
			return []string{p.Provider, formatBool(p.Active), p.ApplicationID, p.IndexName}
		},
	}
}

// newIndexingCommand groups the Indexing Service commands: the index
// providers, which it reads, creates, updates and deletes, and reindexing,
// which it starts and follows.
func newIndexingCommand(a *app) *cobra.Command {
	storefront := resource[indexProvider]{
		use:  "public",
		noun: "index providers as a storefront sees them",
		long: "Reads the public index provider configurations, as a storefront sees them. The search key\n" +
			"is only in -o json.",
		path: "/indexing/{tenant}/public/configurations",
		read: (*emporix.Client).GetUnpaged,
		view: providerView("public provider"),
	}
	providers := resource[indexProvider]{
		use:   "providers",
		noun:  "index providers",
		short: "Read, create, update and delete index providers",
		long: "Reads index provider configurations. Search and write keys are only in -o json;\n" +
			"`epx indexing providers public` shows what a storefront sees and needs no scope.",
		path:   "/indexing/{tenant}/configurations",
		scopes: []string{indexingScope},
		read:   (*emporix.Client).GetUnpaged,
		view:   providerView("provider"),
	}
	return newGroup(indexingUse, "Read and configure search index providers; reindex",
		"Reads and configures the search index providers (e.g. Algolia) and starts and follows\n"+
			"reindex jobs.",
		newResourceCommand(a, providers, newResourceCommand(a, storefront),
			providerCreate(providers).command(a), providerUpdate(providers).command(a),
			itemDelete(providers, "Delete an index provider configuration",
				"Deletes the configuration of PROVIDER, as `epx indexing providers list` shows it.",
				[]string{indexingManageScope}).command(a)),
		newReindexJobsCommand(a),
		reindexProducts().command(a),
	)
}

// providerFields says what the body of a provider configuration holds.
const providerFields = "The JSON in --file holds provider (ALGOLIA or BATTERY_INCLUDED, required), active,\n" +
	"applicationId, indexName, searchKey, writeKey and, for BATTERY_INCLUDED, excludedMixinKeys\n" +
	"or includedMixinPaths. Only one provider can be active at a time. Emporix checks the keys\n" +
	"of BATTERY_INCLUDED before it stores them: it refuses invalid ones (400 Bad Request, exit 1)\n" +
	"and stores nothing when it cannot check them (502 Bad Gateway)."

// providerCreate creates the configuration of an index provider in
// providers, from --file.
func providerCreate(providers resource[indexProvider]) *write {
	return &write{
		use:     createUse,
		short:   "Create an index provider configuration",
		long:    "Creates the configuration of an index provider, whose ID is the provider's name.\n" + providerFields,
		example: "  epx indexing providers create --file algolia.json --dry-run",
		method:  http.MethodPost,
		req:     request{path: providers.path},
		scopes:  []string{indexingManageScope},
		body:    &payload{},
	}
}

// providerUpdate replaces the configuration of an index provider in
// providers, from --file.
func providerUpdate(providers resource[indexProvider]) *write {
	return &write{
		use:   providers.itemUse("update"),
		short: "Replace an index provider configuration",
		long: "Replaces the configuration of PROVIDER, as `epx indexing providers list` shows it. The\n" +
			"reference documents not found (exit 4) for an unknown one, as update creates none; for\n" +
			"BATTERY_INCLUDED Emporix checks the write key first, so an invalid configuration is answered\n" +
			"400 before that (the not-found answer was not checked with a valid key).\n" + providerFields,
		example: "  epx indexing providers update ALGOLIA --file algolia.json",
		method:  http.MethodPut,
		req:     request{path: providers.itemPath()},
		scopes:  []string{indexingManageScope},
		body:    &payload{},
		ids:     providers.itemIDs(),
	}
}

// reindexProducts reindexes every product, as Emporix did before reindex
// jobs, which it deprecated the endpoint for.
func reindexProducts() *write {
	return &write{
		use:   "reindex",
		short: "Reindex every product (deprecated by Emporix)",
		long: "Reindexes every product of the tenant, sending the mode FULL, the only one there is, unless\n" +
			"--file gives the body. Emporix has deprecated this endpoint: `epx indexing reindex-jobs create`\n" +
			"starts a reindex job instead, which its --wait follows. For BATTERY_INCLUDED, Emporix checks\n" +
			"the stored keys first (400 Bad Request when they are invalid, 502 when it cannot check them).",
		example: "  epx indexing reindex --dry-run",
		method:  http.MethodPost,
		req:     request{path: "/indexing/{tenant}/reindex"},
		scopes:  []string{indexingManageScope},
		body:    &payload{typed: objectBody{fixedField("mode", "FULL")}},
	}
}

// newReindexJobsCommand reads reindex jobs and starts them.
func newReindexJobsCommand(a *app) *cobra.Command {
	jobs := reindexJobs()
	return newResourceCommand(a, jobs, reindexJobCreate(jobs).command(a))
}

// reindexJobs is the collection of reindex jobs. The endpoint answers 500
// when asked for its total, so list pages without X-Total-Count.
func reindexJobs() resource[reindexJob] {
	return resource[reindexJob]{
		use:   "reindex-jobs",
		under: indexingUse,
		short: "Read and start reindex jobs",
		long: "Reads reindex jobs and their progress. Emporix fails when asked for the number of jobs,\n" +
			"so list cannot check it: a job that moves between pages while they are read can be shown\n" +
			"twice or missed.",
		path:    "/indexing/{tenant}/reindex-jobs",
		filters: listFilters{kinds: filterable},
		scopes:  []string{indexingScope},
		read:    (*emporix.Client).GetAllWithoutTotal,
		view: view[reindexJob]{
			singular: "reindex job",
			header: []string{
				columnID, "ENTITY", columnStatus, "PROCESSED", "EXPECTED", columnCreated, columnModified, columnMessage,
			},
			row: func(j reindexJob, _ language) []string {
				return []string{
					j.ID, j.EntityType, j.Status, j.ProcessedCount.String(), j.ExpectedCount.String(),
					j.Metadata.CreatedAt, j.Metadata.ModifiedAt, j.Message,
				}
			},
		},
	}
}

// state returns what --wait follows of j: how many entities it processed of
// those it expects, where it says both (one that ended with SUCCESS says only
// how many it expected), and its message, why it failed.
func (j *reindexJob) state() jobState {
	s := jobState{ID: j.ID, Status: j.Status, Failure: j.Message}
	if j.ProcessedCount != "" && j.ExpectedCount != "" {
		s.Progress = j.ProcessedCount.String() + " of " + j.ExpectedCount.String()
	}
	return s
}

// reindexJobCreate starts a reindex job in jobs, which --wait follows as
// the get of jobs reads it.
func reindexJobCreate(jobs resource[reindexJob]) *write {
	return &write{
		use:   createUse,
		short: "Start a reindex job",
		long: "Starts a job that indexes every entity of a type again: PRODUCT, ORDER or a custom entity\n" +
			"type (`epx custom-entities list`); with --rag, it also embeds products for RAG search.\n" +
			"Emporix answers with the job, or with the job of that type that is IN_PROGRESS already.\n" +
			"A job is PENDING while others take its turn, PUBLISHED once every entity was sent to the\n" +
			"index and SUCCESS once the index took them all. For BATTERY_INCLUDED, Emporix checks the\n" +
			"stored keys before it starts a PRODUCT job (400 Bad Request when they are invalid, 502 when\n" +
			"it cannot check them).",
		example: "  epx indexing reindex-jobs create --entity-type PRODUCT --rag --wait\n" +
			"  epx indexing reindex-jobs create --file reindex.json --yes",
		method: http.MethodPost,
		req:    request{path: jobs.path},
		scopes: []string{indexingManageScope},
		body: &payload{typed: objectBody{
			stringField("entityType", "entity-type",
				"the `TYPE` of the entities to reindex: PRODUCT, ORDER or a custom entity type").require(),
			switchField("rag", "rag", "embed products for RAG search too (PRODUCT only)"),
		}},
		answer: followJob(jobs, (*reindexJob).state, []string{"PENDING", "IN_PROGRESS", "PUBLISHED"}, "SUCCESS"),
	}
}
