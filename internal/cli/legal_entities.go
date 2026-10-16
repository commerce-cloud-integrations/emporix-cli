package cli

import (
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// legalEntity is the part of a Customer Management legal entity shown in tables.
type legalEntity struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	ParentID     string      `json:"parentId"`
	AccountLimit valueAmount `json:"accountLimit"`
	LegalInfo    struct {
		LegalName             string `json:"legalName"`
		CountryOfRegistration string `json:"countryOfRegistration"`
	} `json:"legalInfo"`
}

// newLegalEntitiesCommand reads legal entities, the companies of B2B
// customers and their subsidiaries, also those a search finds, and the
// parents of one; and creates, upserts and deletes them.
func newLegalEntitiesCommand(a *app) *cobra.Command {
	entities := resource[legalEntity]{
		use:   "legal-entities",
		short: "Read and change legal entities (B2B companies); read their parents",
		about: "Reads and changes legal entities, the companies of B2B customers and their subsidiaries,\n" +
			"personal data included.\n" + versionedWritesNote,
		long: "Reads legal entities, the companies of B2B customers and their subsidiaries, personal data\n" +
			"included (approval groups name customers). TYPE is COMPANY or SUBSIDIARY, PARENT the legal\n" +
			"entity a subsidiary belongs to, COUNTRY the country of registration. Registration details,\n" +
			"addresses, customer groups, approval groups and restrictions are only in -o json.",
		path:    "/customer-management/{tenant}/legal-entities",
		scopes:  []string{"customermanagement.legalentity_read"},
		filters: listFilters{kinds: filterable},
		view:    legalEntityView(),
	}
	reads := []*cobra.Command{childCommand(a, entities, child[legalEntity]{
		use:   "hierarchy",
		sub:   "parent-hierarchy",
		short: "List a legal entity and its parents",
		long: "Reads a legal entity and every one above it, up to its company, in the order Emporix\n" +
			"sends them, personal data included; the columns are those of `epx legal-entities list`.",
		fetch:   listing{read: (*emporix.Client).GetUnpaged},
		unknown: listsNothing,
		view:    legalEntityView(),
	}), entities.searchCommand(a,
		"Reads the legal entities, personal data included, that an Emporix q filter finds (--query;\n"+
			"without it, every legal entity).", queryField())}
	writes := entities.versionedWrites(a, []string{"customermanagement.legalentity_manage"}, &versioned{
		members: legalEntityMembers, required: "name and type",
		createNote: "A create that Emporix refuses as a conflict may still make the IAM customer groups of\n" +
			"the legal entity.",
		updateNote: "Keep every customer group that Emporix made for the legal entity (Admin, Buyer, Requester\n" +
			"and Contact) in customerGroups: without the Contact group, Emporix deletes that group from\n" +
			"IAM; without another, it leaves that group in IAM, where deleting the legal entity no\n" +
			"longer removes it.",
		deleteNote: "The assignments are the contact assignments of it and of its subsidiaries and the segment\n" +
			"assignments that name it; the IAM customer groups those still in its customerGroups. Its\n" +
			"customer locations stay, and its customers no longer name it (one with no other legal\n" +
			"entity becomes a B2C customer).",
		unknown: answeredDeleted,
		cascade: "the legal entity with its subsidiaries, assignments and IAM customer groups",
	})
	return newResourceCommand(a, entities, slices.Concat(reads, writes)...)
}

// legalEntityMembers are the members of the body of a legal entity's create.
const legalEntityMembers = "name (required), type (COMPANY, the default, or SUBSIDIARY), parentId (the legal\n" +
	"entity a subsidiary belongs to), legalInfo, accountLimit, customerGroups (IAM groups),\n" +
	"entitiesAddresses and approvalGroup (its customer locations and contact assignments),\n" +
	"restrictions and mixins; parentId, customerGroups, entitiesAddresses and approvalGroup must\n" +
	"name items that exist (400 Bad Request, exit 1, otherwise), the last three each as\n" +
	"{\"id\": ...} or in full, as get prints it."

// legalEntityView is the table of legal entities.
func legalEntityView() view[legalEntity] {
	return view[legalEntity]{
		singular: "legal entity",
		header:   []string{columnID, columnName, columnType, columnParent, "LEGAL NAME", columnCountry, "ACCOUNT LIMIT"},
		row: func(e legalEntity, _ language) []string {
			return []string{
				e.ID, e.Name, e.Type, e.ParentID, e.LegalInfo.LegalName, e.LegalInfo.CountryOfRegistration,
				e.AccountLimit.String(),
			}
		},
	}
}
