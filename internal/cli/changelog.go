package cli

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// changeEntry is the part of a Changelog Service entry shown in tables.
type changeEntry struct {
	At       string                     `json:"at"`
	Type     string                     `json:"type"`
	Entity   string                     `json:"entity"`
	EntityID string                     `json:"entityId"`
	Name     string                     `json:"name"`
	Actor    string                     `json:"actor"`
	Paths    map[string]json.RawMessage `json:"paths"` // only the changed fields' names are shown
}

// changedFields returns the names of the changed fields, sorted (a long
// value, which tables cut); -o json shows them with their values.
func changedFields(paths map[string]json.RawMessage) string {
	return strings.Join(slices.Sorted(maps.Keys(paths)), ", ")
}

// newChangelogCommand reads the tenant's change history (audit log).
func newChangelogCommand(a *app) *cobra.Command {
	changes := resource[changeEntry]{
		use:  "changelog",
		noun: "changes",
		long: "Reads the tenant's change history, an audit log (personal data: who changed what, with\n" +
			"the old and new values in -o json), ordered by when the changes occurred. AT is when, TYPE\n" +
			"create, update or delete, ENTITY and ENTITY ID the changed document (an order, customer,\n" +
			"product, site, …), ACTOR who changed it (a name, system, external or unknown) and FIELDS\n" +
			"the changed fields. Without a lower bound on occurredAt (not only inside OR), Emporix\n" +
			"returns the last 30 days, a window that moves while the pages are read: on an active\n" +
			"tenant, read a closed range in the past, e.g. --query\n" +
			"'occurredAt:(>\"2026-09-01T00:00:00.000Z\" AND <\"2026-10-01T00:00:00.000Z\")', or use\n" +
			"--limit; otherwise a change can come or go between pages, so the read fails or misses\n" +
			"it. One document's history: --query 'entity:order entityId:ID'.",
		path:    "/changelog/{tenant}/changelogs",
		scopes:  []string{"changelog.changelog_read"},
		filters: listFilters{kinds: queryable},
		read:    (*emporix.Client).GetAllPageObjects,
		view: view[changeEntry]{
			singular: "change",
			header:   []string{"AT", columnType, "ENTITY", "ENTITY ID", columnName, "ACTOR", columnFields},
			long:     []string{columnFields},
			row: func(c changeEntry, _ language) []string {
				return []string{c.At, c.Type, c.Entity, c.EntityID, c.Name, c.Actor, changedFields(c.Paths)}
			},
		},
	}
	return newGroup(changes.use, "Read the change history (audit log)", changes.long, changes.listCommand(a))
}
