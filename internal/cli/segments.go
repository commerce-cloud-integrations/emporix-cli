package cli

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/spf13/cobra"
)

// segment is the part of a Customer Segment Service segment shown in tables.
type segment struct {
	ID       string    `json:"id"`
	Name     localized `json:"name"`
	Status   string    `json:"status"`
	SiteCode string    `json:"siteCode"`
	Validity validity  `json:"validity"`
}

// segmentCustomer is a customer, of a legal entity for B2B, assigned to a segment.
type segmentCustomer struct {
	Customer struct {
		ID string `json:"id"`
		personName
		Email string `json:"email"`
	} `json:"customer"`
	LegalEntity struct {
		ID string `json:"id"`
	} `json:"legalEntity"`
	AssignmentType string `json:"assignmentType"`
}

// segmentGroup is an IAM group assigned to a segment.
type segmentGroup struct {
	Group struct {
		ID   string    `json:"id"`
		Name localized `json:"name"`
	} `json:"group"`
}

// segmentItem is a product or category assigned to a segment.
type segmentItem struct {
	SegmentID string `json:"segmentId"`
	Type      string `json:"type"`
	Item      struct {
		ID   string    `json:"id"`
		Code string    `json:"code"`
		Name localized `json:"name"`
	} `json:"item"`
}

// segmentMatchItem is a product of a segment match: one that the request
// names to check, and one that the answer lists as matching.
type segmentMatchItem struct {
	ID string `json:"id"`
}

// newSegmentsCommand reads and changes customer segments and what is
// assigned to them (customers, IAM groups and items), all or one, also those
// a search finds, and reads which products belong to segments, which Emporix
// answers only with a write-capable scope (opt-in).
func newSegmentsCommand(a *app) *cobra.Command {
	segments := resource[segment]{
		use:   "segments",
		short: "Read and change customer segments and what is assigned to them",
		about: "Reads and changes customer segments, which group customers to offer them items such as\n" +
			"products or categories, and what is assigned to them: customers, IAM groups and items.",
		long: "Reads customer segments, which group customers to offer them items such as products or\n" +
			"categories. STATUS is ACTIVE or INACTIVE, and FROM and TO the validity window.",
		path:    "/customer-segment/{tenant}/segments",
		scopes:  []string{"customersegment.segment_read"},
		filters: listFilters{kinds: filterable},
		view: view[segment]{
			singular:  "segment",
			header:    []string{columnID, columnName, columnStatus, columnSite, columnFrom, columnTo},
			localized: true,
			row: func(s segment, lang language) []string {
				return []string{s.ID, s.Name.in(lang), s.Status, s.SiteCode, s.Validity.From, s.Validity.To}
			},
		},
	}
	reads := []*cobra.Command{
		segments.searchCommand(a, "Reads the customer segments that an Emporix q filter finds (--query; without it,\n"+
			"every segment).", queryField()),
		newSegmentCustomersCommand(a, segments),
		newSegmentGroupsCommand(a, segments),
		newSegmentItemsCommand(a, segments),
		reader[segmentItem]{view: segmentItems().view, scopes: segments.scopes}.command(a, &command{
			use:   "all-items",
			short: "List the items of every segment",
			long: "Reads the items assigned to every segment, products or categories (TYPE), with their\n" +
				"SEGMENT.",
			req:   request{path: segments.path + "/items"},
			fetch: listing{filters: listFilters{kinds: filterable}},
		}),
		newSegmentMatchCommand(a, segments.path),
	}
	return newResourceCommand(a, segments, slices.Concat(reads, segmentWrites(a, segments))...)
}

// newSegmentCustomersCommand returns the group of the customers of a segment
// of segments: their reads and the writes that assign and remove them.
func newSegmentCustomersCommand(a *app, segments resource[segment]) *cobra.Command {
	customers := segmentCustomers()
	in := segments.nested(customersCommand)
	return newChildGroup(a, in, customers, &childGroup{
		short: "Read and change the customers of a segment",
		long: "Reads the customers assigned to a segment, personal data included, and assigns and removes\n" +
			"them, for B2B with their legal entity.",
		search: &childSearch{short: "Search the customers of a segment", fields: []bodyField{queryField()},
			long: "Reads the customers of a segment that an Emporix q filter finds (--query; without it,\n" +
				"every one), personal data included.\n" + segmentCustomerOrder},
		items: []*childItem{
			{
				use: getUse, short: "Show one customer of a segment",
				long: "Reads one customer assigned to a segment without a legal entity (B2C), personal data\n" +
					"included; one assigned with a legal entity (B2B) is not found (exit 4): get-b2b reads it.",
			},
			{
				use: "get-b2b", short: "Show one customer of a legal entity in a segment",
				long:    "Reads one customer assigned to a segment with a legal entity (B2B), personal data included.",
				columns: []string{columnCustomer, columnLegalEntity},
			},
		},
		writes: segmentCustomerWrites(a, in, customers),
	})
}

// newSegmentGroupsCommand returns the group of the IAM groups of a segment
// of segments: their reads and the writes that assign and remove them.
func newSegmentGroupsCommand(a *app, segments resource[segment]) *cobra.Command {
	groups := segmentGroups()
	in := segments.nested(groupsUse)
	return newChildGroup(a, in, groups, &childGroup{
		short: "Read and change the IAM groups of a segment",
		long:  "Reads the IAM groups assigned to a segment, and assigns and removes them.",
		search: &childSearch{short: "Search the IAM groups of a segment", fields: []bodyField{queryField()},
			long: "Reads the IAM groups assigned to a segment that an Emporix q filter finds (--query;\n" +
				"without it, every one)."},
		items: []*childItem{
			{use: getUse, short: "Show one IAM group of a segment", long: "Reads one IAM group assigned to a segment."},
		},
		writes: segmentGroupWrites(a, in, groups),
	})
}

// newSegmentItemsCommand returns the group of items, the products and
// categories of a segment of segments: their reads and the writes that
// assign and remove them.
func newSegmentItemsCommand(a *app, segments resource[segment]) *cobra.Command {
	items := segmentItems()
	in := segments.nested(itemsUse)
	return newChildGroup(a, in, items, &childGroup{
		short: "Read and change the items of a segment",
		long:  "Reads the items assigned to a segment, products or categories, and assigns and removes them.",
		search: &childSearch{short: "Search the items of a segment", fields: []bodyField{queryField()},
			long: "Reads the items assigned to a segment, products or categories (TYPE), that an Emporix q\n" +
				"filter finds (--query; without it, every one)."},
		items: []*childItem{{
			use: getUse, short: "Show one item of a segment",
			long:    "Reads one item assigned to a segment: " + segmentItemTypeSentence() + ".",
			columns: []string{columnType, columnID}, choices: segmentItemTypes(),
		}},
		writes: segmentItemWrites(a, in, items),
	})
}

// segmentItemTypes are the only values of the TYPE of a segment's item, in
// capitals: Emporix answers others with 400 Bad Request.
func segmentItemTypes() map[string][]string {
	return map[string][]string{columnType: {productItemType, "CATEGORY"}}
}

// segmentItemTypeSentence says which values the TYPE of a segment's item
// takes (segmentItemTypes).
func segmentItemTypeSentence() string {
	return columnType + " is " + alternatives(segmentItemTypes()[columnType]) + ", in capitals"
}

// segmentCustomerOrder says, in the help of the reads of the customers of a
// segment, in which order they are read without --sort: Emporix pages them
// in a different order per request without a sort, repeating some customers
// and leaving out others across pages, so the reads send one
// (segmentCustomers), phrased as the IAM lists say theirs (iamPagingNote).
const segmentCustomerOrder = "list and search sort by id unless --sort says otherwise, an order that Emporix keeps\n" +
	"across pages: unsorted, it pages them in a different order per request."

// segmentCustomers declares the customers of a segment, for B2B with their
// legal entity, read in a stable order: by id unless --sort gives another
// (segmentCustomerOrder), as a user's sort replaces the read's own.
func segmentCustomers() child[segmentCustomer] {
	return child[segmentCustomer]{
		use:   listUse,
		sub:   customersCommand,
		short: "List the customers of a segment",
		long: "Reads the customers of a segment (for B2B, with their legal entity), personal data included.\n" +
			segmentCustomerOrder,
		listQuery: url.Values{sortParameter: {"id"}},
		fetch:     listing{filters: listFilters{kinds: filterable}},
		unknown:   listsNothing,
		view: view[segmentCustomer]{
			singular: "segment customer",
			header:   []string{columnCustomer, columnName, columnEmail, columnLegalEntity, "ASSIGNMENT"},
			legend: "CUSTOMER is the customer's id, not the NUMBER that `epx customers get` takes; ASSIGNMENT\n" +
				"is MANUAL or AUTOMATIC.",
			row: func(c segmentCustomer, _ language) []string {
				return []string{
					c.Customer.ID, c.Customer.fullName(), c.Customer.Email, c.LegalEntity.ID, c.AssignmentType,
				}
			},
		},
	}
}

// segmentGroups declares the IAM groups of a segment.
func segmentGroups() child[segmentGroup] {
	return child[segmentGroup]{
		use:     listUse,
		sub:     groupsUse,
		short:   "List the IAM groups of a segment",
		long:    "Reads the IAM groups assigned to a segment.",
		fetch:   listing{filters: listFilters{kinds: filterable}},
		unknown: listsNothing,
		view: view[segmentGroup]{
			singular:  "segment IAM group",
			header:    []string{columnGroup, columnName},
			localized: true,
			row:       func(g segmentGroup, lang language) []string { return []string{g.Group.ID, g.Group.Name.in(lang)} },
		},
	}
}

// segmentItems declares the items of a segment, products or categories.
func segmentItems() child[segmentItem] {
	return child[segmentItem]{
		use:     listUse,
		sub:     itemsUse,
		short:   "List the items of a segment",
		long:    "Reads the items assigned to a segment, products or categories (TYPE).",
		fetch:   listing{filters: listFilters{kinds: filterable}},
		unknown: listsNothing,
		view: view[segmentItem]{
			singular:  "segment item",
			header:    []string{columnID, "SEGMENT", columnType, columnCode, columnName},
			localized: true,
			row: func(i segmentItem, lang language) []string {
				return []string{i.Item.ID, i.SegmentID, i.Type, i.Item.Code, i.Item.Name.in(lang)}
			},
		},
	}
}

// newSegmentMatchCommand returns `match`, which reads the products of a list
// that belong to segments: a POST to path/match that only reads. Its answer,
// the products named that match, is read as one response, not as a
// collection: a product named twice may be listed twice.
func newSegmentMatchCommand(a *app, path string) *cobra.Command {
	product := func(id string) (segmentMatchItem, error) { return segmentMatchItem{ID: id}, nil }
	return reader[segmentMatchItem]{view: view[segmentMatchItem]{
		singular: "matched product",
		header:   []string{columnProduct},
		row:      func(p segmentMatchItem, _ language) []string { return []string{p.ID} },
	}, scopes: []string{segmentManageScope}}.command(a, &command{
		use:   "match",
		short: "List the products that belong to segments",
		long: "Reads which of the products that --product names (the product id, as in `epx products\n" +
			"list`) belong to an active segment that --segment names, on the site that --site names,\n" +
			"directly or through a category. Emporix answers them whole, in one response: a product\n" +
			"named twice is listed twice, and unknown products are left out.",
		req: request{path: path + "/match"},
		fetch: bodyRead{method: http.MethodPost, body: objectBody{
			{member: "items", value: objectsValue("product", "a product `ID` to check (repeatable)", product).require()},
			{member: "segmentIds", value: listValue("segment", "a segment `ID` to check against (repeatable)").require()},
			siteCodeField(siteCodeParam, "the site `code` to match on, e.g. main").require(),
		}, fetch: single{items: listItems}},
	})
}
