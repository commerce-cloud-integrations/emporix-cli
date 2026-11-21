package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
)

func TestNestedChildSearchAndWrites(t *testing.T) {
	stores := resource[thing]{use: "stores", under: "things", path: "/stores/{site}", view: thingView()}
	shelvesOf := stores.nested("shelves")
	listed := listedChildOf(listing{}, child[thing]{use: listUse, sub: shelvesOf.use, view: thingView()})
	shelves := listed.child
	search := childSearchCommand(&app{}, shelvesOf, listed, &childSearch{short: "Search shelves",
		long: "Reads the shelves a q filter finds.", fields: []bodyField{queryField()}})
	const help = "It reads and shows them as `epx things stores shelves list THING_ID --site S` does."
	if search.Use != "search THING_ID" || !strings.Contains(search.Long, help) {
		t.Errorf("search %q, help %q; want search THING_ID and %q", search.Use, search.Long, help)
	}
	for _, tt := range []struct {
		at, want  collectionAt
		uses      [2]string
		completes []string
	}{
		{
			at: childAt(shelvesOf, shelves),
			want: collectionAt{path: "/stores/{site}/{parent}/shelves", parent: "THING_ID", item: "ID",
				itemPath: "/stores/{site}/{parent}/shelves/{id}", get: "things stores shelves get THING_ID ID --site S"},
			uses:      [2]string{"create THING_ID", "update THING_ID ID"},
			completes: []string{"ID", "THING_ID"},
		},
		{
			at: stores.tenantCollection(),
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

func TestChildNamesItsPath(t *testing.T) {
	defer func() {
		if got, want := fmt.Sprint(recover()), "shelves: a child names its path below the parent item in sub"; got != want {
			t.Fatalf("panic %q; want %q", got, want)
		}
	}()
	stores := resource[thing]{use: "stores", path: "/stores", view: thingView()}
	childCommand(&app{}, stores, child[thing]{use: "shelves", fetch: listing{}, view: thingView()})
}

func TestChildGroupNeedsANestedParent(t *testing.T) {
	defer func() {
		if got := recover(); got != "thing: a child in a group of its own is below a nested parent" {
			t.Fatalf("recovered %v; want the declaration named", got)
		}
	}()
	stores := resource[thing]{use: "stores", path: "/stores", view: thingView()}
	newChildGroup(&app{}, stores, child[thing]{use: listUse, sub: "shelves", fetch: listing{}, view: thingView()},
		&childGroup{})
}

func TestChildOwnScopesNameOne(t *testing.T) {
	defer func() {
		if got, want := fmt.Sprint(recover()),
			"a child's own scopes name one at least: one that requests none declares noScopes"; got != want {
			t.Fatalf("panic %q; want %q", got, want)
		}
	}()
	child[thing]{scopes: ownScopes{}}.reader([]string{"thing.read"})
}

func TestListedChildListsAsItsSearchReads(t *testing.T) {
	members := listing{filters: listFilters{kinds: filterable}}
	listed := listedChildOf(members, child[thing]{sub: "shelves", view: thingView()})
	if got, ok := listed.fetch.(listing); !ok || got.filters.kinds != members.filters.kinds {
		t.Fatalf("fetch %#v; want the listing its search reads", listed.fetch)
	}
}
