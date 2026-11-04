package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
)

// TestNestedChildSearchAndWrites checks a child collection listed by `list`
// below a group of its own (resource.nested), as one with writes is: its
// search is named search, takes the parent's argument and names the list as
// that group's; its writes (childAt) send below one item of the parent,
// whose argument they take first, and name one item as get does; and those
// of a resource (resource.at) send to its collection. Both complete the
// arguments of one item.
func TestNestedChildSearchAndWrites(t *testing.T) {
	stores := resource[thing]{use: "stores", under: "things", path: "/stores/{site}", view: thingView()}
	shelvesOf := stores.nested("shelves")
	shelves := child[thing]{use: listUse, sub: shelvesOf.use, fetch: listing{}, view: thingView()}
	search := childSearchCommand(&app{}, shelvesOf, shelves, "Search shelves", "Reads the shelves a q filter finds.",
		queryField())
	const help = "It reads and shows them as `epx things stores shelves list THING_ID --site S` does."
	if search.Use != "search THING_ID" || !strings.Contains(search.Long, help) {
		t.Errorf("search %q, help %q; want search THING_ID and %q", search.Use, search.Long, help)
	}
	for _, tt := range []struct {
		at, want  collectionAt
		uses      [2]string // create and update, as at names them
		completes []string  // the arguments of one item that complete
	}{
		{
			at: childAt(shelvesOf, shelves),
			want: collectionAt{path: "/stores/{site}/{parent}/shelves", parent: "THING_ID", item: "ID",
				itemPath: "/stores/{site}/{parent}/shelves/{id}", get: "things stores shelves get THING_ID ID --site S"},
			uses:      [2]string{"create THING_ID", "update THING_ID ID"},
			completes: []string{"ID", "THING_ID"},
		},
		{
			at: stores.at(),
			want: collectionAt{path: "/stores/{site}", item: "ID", itemPath: "/stores/{site}/{id}",
				get: "things stores get ID --site S"},
			uses:      [2]string{"create", "update ID"},
			completes: []string{"ID"},
		},
	} {
		where := func(c collectionAt) [5]string { return [5]string{c.path, c.parent, c.item, c.itemPath, c.get} }
		completes := slices.Sorted(maps.Keys(tt.at.itemIDs))
		if where(tt.at) != where(tt.want) || tt.at.use(createUse) != tt.uses[0] ||
			tt.at.itemUse("update") != tt.uses[1] || !slices.Equal(completes, tt.completes) {
			t.Errorf("%+v, uses %q and %q, completes %q; want %+v, %q and %q", tt.at, tt.at.use(createUse),
				tt.at.itemUse("update"), completes, tt.want, tt.uses, tt.completes)
		}
	}
}

// TestNestedKeepsWhereItsItemsAreListed checks that a resource nested in a
// group of its own (resource.nested) moves its commands below that group and
// still names the command line that lists its items, and that it completes
// their IDs from its own list only where it has one: one that another command
// lists at another path, such as the import streams of a configuration, is
// not claimed to be listed at its path, and completes none.
func TestNestedKeepsWhereItsItemsAreListed(t *testing.T) {
	stores := resource[thing]{use: "stores", under: "things", path: "/stores", view: thingView()}
	elsewhere := stores
	elsewhere.listedBy = "things shops stores SHOP_ID"
	for _, tt := range []struct {
		nested            resource[thing]
		listedBy          string
		atPath, completes bool
	}{
		{nested: stores.nested("shelves"), listedBy: "things stores list", atPath: true, completes: true},
		{nested: elsewhere.nested("shelves"), listedBy: elsewhere.listedBy},
	} {
		r := tt.nested
		if r.under != "things stores" || r.use != "shelves" || r.listedBy != tt.listedBy || r.atPath != tt.atPath ||
			(r.ids() != nil) != tt.completes {
			t.Errorf("nested %+v, completes %v; want below things stores, listed by %q, at its path %v, completes %v",
				r, r.ids() != nil, tt.listedBy, tt.atPath, tt.completes)
		}
	}
}

// TestChildListedByListNamesItsPath checks that a child listed by `list`
// without sub, which would send to .../list, a mistake in its declaration,
// fails while the command tree is built.
func TestChildListedByListNamesItsPath(t *testing.T) {
	defer func() {
		if got, want := fmt.Sprint(recover()), "thing: a child listed by list names its path in sub"; got != want {
			t.Fatalf("panic %q; want %q", got, want)
		}
	}()
	stores := resource[thing]{use: "stores", path: "/stores", view: thingView()}
	childCommand(&app{}, stores.nested("shelves"), child[thing]{use: listUse, fetch: listing{}, view: thingView()})
}
