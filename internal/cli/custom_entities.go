package cli

import (
	"slices"

	"github.com/spf13/cobra"
)

// customEntity is the part of a custom entity type shown in tables.
type customEntity struct {
	ID       string       `json:"id"`
	Name     localized    `json:"name"`
	Metadata itemMetadata `json:"metadata"`
}

// customInstance is the part of a custom entity instance shown in tables.
type customInstance struct {
	ID       string        `json:"id"`
	Name     localized     `json:"name"`
	Owner    instanceOwner `json:"owner"`
	Metadata itemMetadata  `json:"metadata"`
}

// instanceOwner is the owner of a custom instance, by default who created it.
type instanceOwner struct {
	Type   string `json:"type"`
	UserID string `json:"userId"`
}

func (o instanceOwner) String() string { return spaced(o.Type, o.UserID) }

// instancesUse is the group of the reads and writes of custom instances.
const instancesUse = "instances"

// instancesSentence ends the help of custom-entities: where the reads and
// writes of instances are.
const instancesSentence = "`epx custom-entities instances` reads and changes the instances of a type."

// newCustomEntitiesCommand reads and changes the tenant's custom entity types
// and their instances, which need the scopes of their type, and exports and
// imports types with their schemas; the export needs a write-capable scope
// (opt-in).
func newCustomEntitiesCommand(a *app) *cobra.Command {
	entities := customEntities()
	export := exportCommand(a, entities.path, &export{
		short: "Export custom entity types with their schemas",
		long: "Reads an export package of the custom entity types that --type names (the ID that\n" +
			"`epx custom-entities list` shows): each type with its schemas, without metadata, and\n" +
			"none of their instances. Emporix leaves out an ID that names no custom entity type (a\n" +
			"built-in type such as PRODUCT too); if none is left, it answers with an empty package\n" +
			"(exit 0).",
		scopes:    []string{schemaManage},
		selection: listValue(typeParam, "a custom entity type `ID` to export (repeatable)"),
	})
	return newResourceCommand(a, entities, slices.Concat([]*cobra.Command{newCustomInstancesCommand(a, entities),
		export}, customEntityWrites(a, entities))...)
}

// customEntities declares the custom entity types, whose IDs the TYPE
// of another read completes with too.
func customEntities() resource[customEntity] {
	return resource[customEntity]{
		use:   "custom-entities",
		noun:  "custom entity types",
		short: "Read and change custom entity types and their instances",
		long:  "Reads custom entity types: entities a tenant defines.\n" + instancesSentence,
		about: "Reads and changes custom entity types: entities a tenant defines. create, update\n" +
			"and delete change types; export exports types with their schemas (opt-in), which\n" +
			"import imports.\n" + instancesSentence,
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

// newCustomInstancesCommand returns the group of the instances of a custom
// entity type of entities: list, search and get read them, and the writes
// (customInstanceWrites) change them, each with the scope of the type.
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
			legend: "OWNER is an instance's owner, by default who created it (EMPLOYEE, CUSTOMER or SERVICE,\n" +
				"and the user or client id); attribute values (mixins) are only in -o json. epx requests\n" +
				"the scope of the type, not the tenant-wide schema.custominstance_read that Emporix also\n" +
				"accepts.",
			row: func(i customInstance, lang language) []string {
				return []string{i.ID, i.Name.in(lang), i.Owner.String(), i.Metadata.Version.String()}
			},
		},
	}
	reads := []*cobra.Command{
		childCommand(a, typesOf, instances),
		childSearchCommand(a, typesOf, instances, "Search the instances of a custom entity type",
			"Reads the instances of a custom entity type that an Emporix q filter finds (--query;\n"+
				"without it, every one).", queryField()),
		childItemCommand(a, typesOf, instances, &childItem{
			use:   "get",
			short: "Show one instance of a custom entity type",
			long:  "Reads one instance of a custom entity type.",
		}),
	}
	return newGroup(instancesUse, "Read and change the instances of a custom entity type",
		"Reads the instances of a custom entity type, every one (list), those a filter finds\n"+
			"(search) or one (get), and creates, changes and deletes them, one at a time or in bulk.",
		slices.Concat(reads, customInstanceWrites(a, typesOf, instances))...)
}
