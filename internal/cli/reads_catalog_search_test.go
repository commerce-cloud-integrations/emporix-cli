package cli

import "net/http"

// The read cases of the product, category and category-tree searches, which
// Emporix serves as POST with a JSON body but which change nothing (see TestReads).

// catalogSearchReads returns the searches of the Product and Category services.
func catalogSearchReads() []readCase {
	const categoryScope = "category.category_read_unpublished"
	trees := `[{"id":"r1","code":"tools","published":true,"localizedName":{"en":"Tools","nl":"Werkplaats"},
		"subcategories":[{"id":"c2","published":false,"localizedName":{"en":"Saws"}}]}]`
	return []readCase{
		{
			// Without --query, the body's q filters nothing; pages are read as list reads them.
			args:    []string{"products", "search"},
			path:    "/product/acme1/products/search",
			scopes:  []string{"product.product_read_unpublished"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body:    productsBody,
			want: "ID      CODE    TYPE     PUBLISHED  NAME\n" +
				"001234  001234  BASIC    false      Transistors (fr)\n" +
				"ZZ1     ZZ1     BASIC    true       Valve\n" +
				"P2      P2      VARIANT             Plain name\n",
		},
		{
			// --query is sent in the body as given, > included.
			args: []string{"products", "search", "--query", "metadata.modifiedAt:>2026-01-01 published:true",
				"--language", "nl"},
			path:    "/product/acme1/products/search",
			scopes:  []string{"product.product_read_unpublished"},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"metadata.modifiedAt:>2026-01-01 published:true"}`,
			body:    `[{"id":"ZZ1","code":"ZZ1","productType":"BASIC","published":true,"name":{"en":"Valve","nl":"Klep"}}]`,
			want:    "ID   CODE  TYPE   PUBLISHED  NAME\nZZ1  ZZ1   BASIC  true       Klep\n",
		},
		{
			// Unpublished categories are asked for, with the scope that shows them, as list does.
			args:    []string{"categories", "search", "--language", "nl"},
			path:    "/category/acme1/categories/search",
			query:   "showUnpublished=true",
			scopes:  []string{categoryScope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":""}`,
			body:    categoriesBody,
			want: "ID  CODE        TYPE      PUBLISHED  NAME\n" +
				"c1  Transistor  STANDARD  true       Transistoren\n" +
				"c2              STANDARD  false      Plain only\n",
		},
		{
			args:    []string{"categories", "search", "--query", "code:Transistor"},
			path:    "/category/acme1/categories/search",
			query:   "showUnpublished=true",
			scopes:  []string{categoryScope},
			filters: listFilters{kinds: sortable},
			method:  http.MethodPost,
			sent:    `{"q":"code:Transistor"}`,
			body:    `[{"id":"c1","code":"Transistor","type":"STANDARD","published":true,"localizedName":{"en":"Transistors"}}]`,
			want:    "ID  CODE        TYPE      PUBLISHED  NAME\nc1  Transistor  STANDARD  true       Transistors\n",
		},
		{
			// Without --category, every tree (as list, seen live); Emporix answers whole, without paging.
			// The optional list is sent as an empty array, never null.
			args:   []string{"category-trees", "search", "--language", "nl"},
			path:   "/category/acme1/category-trees/search",
			query:  "showUnpublished=true",
			scopes: []string{categoryScope},
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"categoryIds":[]}`,
			body:   trees,
			want:   "ID  CODE   PUBLISHED  NAME\nr1  tools  true       Werkplaats\nc2         false        Saws (en)\n",
		},
		{
			// Each --category adds one id, as given.
			args:   []string{"category-trees", "search", "--category", "c2", "--category", "c 3,4"},
			path:   "/category/acme1/category-trees/search",
			query:  "showUnpublished=true",
			scopes: []string{categoryScope},
			mode:   getWhole,
			method: http.MethodPost,
			sent:   `{"categoryIds":["c2","c 3,4"]}`,
			body:   trees,
			want:   "ID  CODE   PUBLISHED  NAME\nr1  tools  true       Tools\nc2         false        Saws\n",
		},
	}
}
