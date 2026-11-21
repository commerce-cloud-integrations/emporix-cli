package cli

import (
	"slices"

	"github.com/spf13/cobra"
)

type customEntity struct {
	ID       string       `json:"id"`
	Name     localized    `json:"name"`
	Metadata itemMetadata `json:"metadata"`
}

type customInstance struct {
	ID       string        `json:"id"`
	Name     localized     `json:"name"`
	Owner    instanceOwner `json:"owner"`
	Metadata itemMetadata  `json:"metadata"`
}

type instanceOwner struct {
	Type   string `json:"type"`
	UserID string `json:"userId"`
}

func (o instanceOwner) String() string { return spaced(o.Type, o.UserID) }

const instancesUse = "instances"

const instancesSentence = "`epx custom-entities instances` reads and changes the instances of a type."

func newCustomEntitiesCommand(a *app) *cobra.Command {
	entities := customEntities()
	export := exportCommand(a, entities.path, &export{
		short: "Export custom entity types with their schemas",
		long: "Reads an export package of the custom entity types that --type names (the ID that " +
			"`epx custom-entities list` shows): each type with its schemas, without metadata, and " +
			"none of their instances. Emporix leaves out an ID that names no custom entity type (a " +
			"built-in type such as PRODUCT too); if none is left, it answers with an empty package " +
			"(exit 0).",
		scopes:    []string{schemaManage},
		selection: listValue(typeParam, "a custom entity type `ID` to export (repeatable)"),
	})
	return newResourceCommand(a, entities, slices.Concat([]*cobra.Command{newCustomInstancesCommand(a, entities),
		export}, customEntityWrites(a, entities))...)
}

func customEntities() resource[customEntity] {
	return resource[customEntity]{
		use:   "custom-entities",
		noun:  "custom entity types",
		short: "Read and change custom entity types and their instances",
		long:  "Reads custom entity types: entities a tenant defines. " + instancesSentence,
		about: "Reads and changes custom entity types: entities a tenant defines. create, update " +
			"and delete change types; export exports types with their schemas (opt-in), which " +
			"import imports. " + instancesSentence,
		path:    "/schema/{tenant}/custom-entities",
		filters: listFilters{kinds: filterable},
		view: view[customEntity]{
			singular:  "custom entity",
			header:    []string{columnID, columnName, columnVersion},
			localized: true,
			row: func(e customEntity, lang language) []string {
				return []string{e.ID, e.Name.in(lang), e.Metadata.Version.String()}
			},
		},
	}
}

func newCustomInstancesCommand(a *app, entities resource[customEntity]) *cobra.Command {
	typesOf := entities.nested(instancesUse)
	instances := child[customInstance]{
		use:    listUse,
		sub:    typesOf.use,
		short:  "List the instances of a custom entity type",
		long:   "Reads the instances of a custom entity type (`epx custom-entities instances get` shows one).",
		scopes: []string{customTypeReadScope},
		fetch:  listing{filters: listFilters{kinds: filterable}},
		view: view[customInstance]{
			singular:  "custom instance",
			header:    []string{columnID, columnName, "OWNER", columnVersion},
			localized: true,
			siteAware: true,
			legend: "OWNER is an instance's owner, by default who created it (EMPLOYEE, CUSTOMER or SERVICE, " +
				"and the user or client id). epx requests the scope of the type, not the tenant-wide " +
				"schema.custominstance_read that Emporix also accepts.",
			row: func(i customInstance, lang language) []string {
				return []string{i.ID, i.Name.in(lang), i.Owner.String(), i.Metadata.Version.String()}
			},
		},
	}
	reads := []*cobra.Command{
		childCommand(a, typesOf, instances),
		childSearchCommand(a, typesOf, instances, "Search the instances of a custom entity type",
			"Reads the instances of a custom entity type that an Emporix q filter finds (--query; "+
				"without it, every one).", queryField()),
		childItemCommand(a, typesOf, instances, &childItem{
			use:   "get",
			short: "Show one instance of a custom entity type",
			long:  "Reads one instance of a custom entity type.",
		}),
	}
	return newGroup(instancesUse, "Read and change the instances of a custom entity type",
		"Reads the instances of a custom entity type, every one (list), those a filter finds "+
			"(search) or one (get), and creates, changes and deletes them, one at a time or in bulk.",
		slices.Concat(reads, customInstanceWrites(a, typesOf, instances))...)
}
