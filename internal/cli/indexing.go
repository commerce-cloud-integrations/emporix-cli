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
		long: "Reads the public index provider configurations, as a storefront sees them. The search key " +
			"is only in `-o json`.",
		path: "/indexing/{tenant}/public/configurations",
		read: (*emporix.Client).GetUnpaged,
		view: providerView("public provider"),
	}
	providers := resource[indexProvider]{
		use:   "providers",
		noun:  "index providers",
		short: "Read, create, update and delete index providers",
		long: "Reads index provider configurations. Search and write keys are only in `-o json`, as are " +
			"the BatteryIncluded settings apiUrl, enrichments and siteAwareMixinsStrategy, where they are set " +
			"(the default applies otherwise); " +
			"`epx indexing providers public` shows what a storefront sees and needs no scope.",
		path:   "/indexing/{tenant}/configurations",
		scopes: []string{indexingScope},
		read:   (*emporix.Client).GetUnpaged,
		view:   providerView("provider"),
	}
	return newGroup(indexingUse, "Read and configure search index providers; reindex",
		"Reads and configures the search index providers (e.g. Algolia) and starts and follows "+
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

// batteryIncludedSettings says what the BatteryIncluded-only settings of a
// provider configuration hold, when Emporix refuses them and when a change
// reaches the index.
const batteryIncludedSettings = "For BATTERY_INCLUDED only, it may hold excludedMixinKeys or " +
	"includedMixinPaths, and: apiUrl, the BatteryIncluded API that holds the index, an https address of " +
	"a host that Emporix allows, " +
	"without path, query, fragment, port or user (the default host without it); enrichments, the Emporix " +
	"services that product indexing calls besides Product, Site Settings and Configuration (AVAILABILITY, " +
	"PRICE and CUSTOMER_SEGMENT add their data to the indexed document; FEE calls the Fee Service but adds " +
	"nothing to it), the first three without it and none for an empty list; and siteAwareMixinsStrategy, " +
	"how site-aware mixins are indexed (FAN_OUT, the default, or INHERITANCE). Emporix refuses apiUrl, " +
	"enrichments (even an empty list) and siteAwareMixinsStrategy for ALGOLIA, and an apiUrl of another " +
	"host (400 Bad Request, exit 1). A change of any of the three takes effect within 10 minutes; a new " +
	"enrichments or siteAwareMixinsStrategy reaches a product when it is next updated or reindexed."

// providerFields says what the body of a provider configuration holds.
const providerFields = "The JSON in --file holds provider (ALGOLIA or BATTERY_INCLUDED, required), active, " +
	"applicationId, indexName, searchKey and writeKey. " + batteryIncludedSettings + " Only one provider " +
	"can be active at a time. Emporix checks the keys of BATTERY_INCLUDED before it stores them, against " +
	"the host of apiUrl: it refuses invalid ones (400 Bad Request, exit 1) and stores nothing when it " +
	"cannot check them (502 Bad Gateway)."

// providerCreate creates the configuration of an index provider in
// providers, from --file.
func providerCreate(providers resource[indexProvider]) *write {
	return &write{
		use:     createUse,
		short:   "Create an index provider configuration",
		long:    "Creates the configuration of an index provider, whose ID is the provider's name. " + providerFields,
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
		long: "Replaces the configuration of PROVIDER, as `epx indexing providers list` shows it. The " +
			"reference documents not found (exit 4) for an unknown one, as update creates none; for " +
			"BATTERY_INCLUDED Emporix checks the write key first, so an invalid configuration is answered " +
			"400 before that (the not-found answer was not checked with a valid key). The configuration " +
			"is replaced whole: each of apiUrl, enrichments and siteAwareMixinsStrategy that the body leaves " +
			"out goes back to its default (for apiUrl the default host, on which the write key must then be " +
			"valid), so start from `epx indexing providers get PROVIDER -o json`. Change apiUrl while no " +
			"product reindex runs; to apply a change of enrichments or siteAwareMixinsStrategy to every product at once, or " +
			"after a change of apiUrl to an index that " +
			"does not hold every product yet, reindex the products once it takes effect " +
			"(`epx indexing reindex-jobs create --entity-type PRODUCT`). " + providerFields,
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
		long: "Reindexes every product of the tenant, sending the mode FULL, the only one there is, unless " +
			"--file gives the body. Emporix has deprecated this endpoint: `epx indexing reindex-jobs create` " +
			"starts a reindex job instead, which its --wait follows. For BATTERY_INCLUDED, Emporix checks " +
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
		long: "Reads reindex jobs and their progress. Emporix fails when asked for the number of jobs, " +
			"so list cannot check it: a job that moves between pages while they are read can be shown " +
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

// entityBuiltInTypes returns the entity types that reindex jobs and the AI
// RAG Indexer take besides the custom ones, in any case.
func entityBuiltInTypes() builtInTypes {
	return anyCaseTypes("PRODUCT", "ORDER", "CUSTOMER", "LEGAL-ENTITY")
}

// entityTypeSentences say, for help, what TYPE, an entity type, is: one of
// builtIn, the entityBuiltInTypes that its caller declares a command with,
// LEGAL-ENTITY with a hyphen, or a custom one.
func entityTypeSentences(builtIn builtInTypes) string {
	return builtIn.clause("TYPE") + ". LEGAL-ENTITY takes a hyphen, not an underscore."
}

// reindexJobCreate starts a reindex job in jobs, which --wait follows as
// the get of jobs reads it.
func reindexJobCreate(jobs resource[reindexJob]) *write {
	types := entityBuiltInTypes()
	return &write{
		use:   createUse,
		short: "Start a reindex job",
		long: "Starts a job that indexes every entity of TYPE again; with --rag, it also embeds products for " +
			"RAG search. " + entityTypeSentences(types) + " " +
			"Searches of customers and legal entities, in the Search Service and by the tools of agents, " +
			"read what a CUSTOMER or LEGAL-ENTITY job indexes. " +
			"Emporix answers with the job, or with the job of that type that is IN_PROGRESS already. " +
			"A job is PENDING while others take its turn, PUBLISHED once every entity was sent to the " +
			"index and SUCCESS once the index took them all. For BATTERY_INCLUDED, Emporix checks the " +
			"stored keys before it starts a PRODUCT job (400 Bad Request when they are invalid, 502 when " +
			"it cannot check them).",
		example: "  epx indexing reindex-jobs create --entity-type PRODUCT --rag --wait\n" +
			"  epx indexing reindex-jobs create --entity-type LEGAL-ENTITY --wait\n" +
			"  epx indexing reindex-jobs create --file reindex.json --yes",
		method: http.MethodPost,
		req:    request{path: jobs.path},
		scopes: []string{indexingManageScope},
		body: &payload{typed: objectBody{
			stringField("entityType", "entity-type", "the `TYPE` of the entities to reindex: "+types.kinds()).
				require().suggesting(types.names()...),
			switchField("rag", "rag", "embed products for RAG search too (PRODUCT only)"),
		}},
		answer: followJob(jobs, (*reindexJob).state, []string{"PENDING", "IN_PROGRESS", "PUBLISHED"}, "SUCCESS"),
	}
}
