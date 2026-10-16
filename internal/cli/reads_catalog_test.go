package cli

// The read cases of the Product, Category, Catalog, Brand, Label, Schema and Unit Handling
// services (see TestReads).

// categoriesBody holds a published category with translations and an unpublished one
// with a plain name only.
const categoriesBody = `[
	{"id":"c1","code":"Transistor","type":"STANDARD","published":true,"name":"Transistors",
	 "localizedName":{"en":"Transistors","nl":"Transistoren"}},
	{"id":"c2","code":null,"type":"STANDARD","published":false,"name":"Plain only"}
]`

// productsBody holds products with and without the translations asked for, and one
// with a plain name.
const productsBody = `[
	{"id":"001234","code":"001234","productType":"BASIC","published":false,"name":{"nl":"Transistor","fr":"Transistors"}},
	{"id":"ZZ1","code":"ZZ1","productType":"BASIC","published":true,"name":{"en":"Valve","nl":"Klep"}},
	{"id":"P2","code":"P2","productType":"VARIANT","name":"Plain name"}
]`

// brandsBody holds a brand with an English and a Dutch name, one with a Dutch
// name only and one with a plain name only.
const brandsBody = `[
	{"id":"b1","name":"Brand_1","localizedName":{"en":"Brand one","nl":"Merk een"}},
	{"id":"b2","name":"NORTHWIND TRADE","localizedName":{"nl":"NORTHWIND TRADE"}},
	{"id":"b3","name":"Plain only"}]`

// catalogsBody holds a visible catalog named in English and one that
// reproduces the live disagreement between status and visibility.
const catalogsBody = `[{"id":"cc","name":{"en":"Main"},
	"status":"visible","visibility":{"visible":true,"from":"2025-10","to":"2028-10"},
	"publishedSites":["NL","main"],"categoryIds":["c1"]},
	{"id":"odd","name":{"nl":"Raar"},"status":"visible","visibility":{"visible":false}}]`

// catalogsTableInDutch is the table of catalogsBody in Dutch.
const catalogsTableInDutch = "ID   NAME       STATUS   VISIBLE  FROM     TO       SITES    ROOT CATEGORIES\n" +
	"cc   Main (en)  visible  true     2025-10  2028-10  NL,main  1\n" +
	"odd  Raar       visible  false                               0\n"

// productReads returns the Product Service reads.
func productReads() []readCase {
	return []readCase{
		{
			args:    []string{"product-templates", "list", "--language", "nl"},
			path:    "/product/acme1/product-templates",
			scopes:  []string{"product.product_template_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"t1","name":{"en":"Transistor","nl":"Transistor NL"},
				"attributes":[{"key":"power","type":"NUMBER"}],"metadata":{"version":3}}]`,
			want: "ID  NAME           ATTRIBUTES  VERSION\nt1  Transistor NL  1           3\n",
		},
		{
			// A missing translation is labeled with the language shown; a plain name is shown as sent.
			args:    []string{"products", "list"},
			path:    "/product/acme1/products",
			scopes:  []string{"product.product_read_unpublished"},
			filters: listFilters{kinds: filterable},
			body:    productsBody,
			want: "ID      CODE    TYPE     PUBLISHED  NAME\n" +
				"001234  001234  BASIC    false      Transistors (fr)\n" +
				"ZZ1     ZZ1     BASIC    true       Valve\n" +
				"P2      P2      VARIANT             Plain name\n",
		},
		{
			// Language tags compare case-insensitively.
			args:    []string{"products", "list", "--language", "NL"},
			path:    "/product/acme1/products",
			scopes:  []string{"product.product_read_unpublished"},
			filters: listFilters{kinds: filterable},
			body:    productsBody,
			want: "ID      CODE    TYPE     PUBLISHED  NAME\n" +
				"001234  001234  BASIC    false      Transistor\n" +
				"ZZ1     ZZ1     BASIC    true       Klep\n" +
				"P2      P2      VARIANT             Plain name\n",
		},
		{
			// The id is one escaped path segment.
			args:   []string{"products", "get", "001 234/x", "--language", "nl"},
			path:   "/product/acme1/products/001%20234%2Fx",
			scopes: []string{"product.product_read_unpublished"},
			mode:   getWhole,
			body: `{"id":"001 234/x","code":"001234","productType":"BASIC","published":false,"name":{"en":"Transistor",
				"nl":"Transistor NL"}}`,
			want: "ID         CODE    TYPE   PUBLISHED  NAME\n001 234/x  001234  BASIC  false      Transistor NL\n",
		},
		{
			args:   []string{"product-templates", "get", "t1", "--language", "nl"},
			path:   "/product/acme1/product-templates/t1",
			scopes: []string{"product.product_template_read"},
			mode:   getWhole,
			body: `{"id":"t1","name":{"en":"Transistor","nl":"Transistor NL"},"attributes":[{"key":"power"}],
				"metadata":{"version":3}}`,
			want: "ID  NAME           ATTRIBUTES  VERSION\nt1  Transistor NL  1           3\n",
		},
		{
			// Opt-in, read with one request: the reference documents no paging. A pending job has
			// not finished; an error is cut.
			args:   []string{"products", "recalculation-jobs", "list", "--allow-write-scope"},
			path:   "/product/acme1/products/recalculate/jobs",
			scopes: []string{"product.product_manage"},
			mode:   getWhole,
			body: `[{"id":"j1","parentRoot":"100200","requestedByIds":["100200-1","100200-2"],"status":"FINISHED",
				"createdAt":"2026-10-13T10:10:00.000Z","processingStartedAt":"2026-10-13T10:10:03.000Z",
				"finishedAt":"2026-10-13T10:10:07.000Z","error":null,"attempts":1},
				{"id":"j2","parentRoot":"100300","status":"FAILED","createdAt":"2026-10-13T10:08:00.000Z",
				"finishedAt":"2026-10-13T10:08:12.000Z","attempts":3,
				"error":"Database write timeout during variant tree update of product 100300 and its variants"},
				{"id":"j3","parentRoot":"100400","status":"PENDING","createdAt":"2026-10-13T10:11:00.000Z",
				"processingStartedAt":null,"finishedAt":null,"error":null,"attempts":0}]`,
			want: "ID  ROOT PRODUCT  STATUS    ATTEMPTS  CREATED                   FINISHED                  ERROR\n" +
				"j1  100200        FINISHED  1         2026-10-13T10:10:00.000Z  2026-10-13T10:10:07.000Z  \n" +
				"j2  100300        FAILED    3         2026-10-13T10:08:00.000Z  2026-10-13T10:08:12.000Z  " +
				"Database write timeout during variant tree update of produc…\n" +
				"j3  100400        PENDING   0         2026-10-13T10:11:00.000Z                            \n",
		},
		{
			args:   []string{"products", "recalculation-jobs", "get", "j/1", "--allow-write-scope"},
			path:   "/product/acme1/products/recalculate/jobs/j%2F1",
			scopes: []string{"product.product_manage"},
			mode:   getWhole,
			body: `{"id":"j/1","parentRoot":"100200","status":"PROCESSING","createdAt":"2026-10-13T10:10:00.000Z",
				"attempts":1}`,
			want: "ID   ROOT PRODUCT  STATUS      ATTEMPTS  CREATED                   FINISHED  ERROR\n" +
				"j/1  100200        PROCESSING  1         2026-10-13T10:10:00.000Z            \n",
		},
	}
}

// categoryReads returns the Category Service reads.
func categoryReads() []readCase {
	return []readCase{
		{
			// Unpublished categories are asked for, with the scope that shows them.
			args:    []string{"categories", "list", "--language", "nl"},
			path:    "/category/acme1/categories",
			query:   "showUnpublished=true",
			scopes:  []string{"category.category_read_unpublished"},
			filters: listFilters{kinds: filterable},
			body:    categoriesBody,
			want: `ID  CODE        TYPE      PUBLISHED  NAME
c1  Transistor  STANDARD  true       Transistoren
c2              STANDARD  false      Plain only
`,
		},
		{
			// A missing translation is labeled; the plain name is only a last resort.
			args:    []string{"categories", "list", "--language", "de"},
			path:    "/category/acme1/categories",
			query:   "showUnpublished=true",
			scopes:  []string{"category.category_read_unpublished"},
			filters: listFilters{kinds: filterable},
			body:    categoriesBody,
			want: `ID  CODE        TYPE      PUBLISHED  NAME
c1  Transistor  STANDARD  true       Transistors (en)
c2              STANDARD  false      Plain only
`,
		},
		{
			args:   []string{"categories", "get", "c2", "--language", "nl"},
			path:   "/category/acme1/categories/c2",
			query:  "showUnpublished=true",
			scopes: []string{"category.category_read_unpublished"},
			mode:   getWhole,
			body: `{"id":"c2","type":"STANDARD","published":false,"name":"Plain only",
				"localizedName":{"nl":"Alleen gewoon"}}`,
			want: "ID  CODE  TYPE      PUBLISHED  NAME\nc2        STANDARD  false      Alleen gewoon\n",
		},
		{
			// The product id is one escaped path segment.
			args:    []string{"categories", "for-product", "001 234/x", "--language", "nl"},
			path:    "/category/acme1/assignments/references/001%20234%2Fx",
			query:   "showUnpublished=true",
			scopes:  []string{"category.category_read_unpublished"},
			filters: listFilters{kinds: sortable},
			body:    categoriesBody,
			want: "ID  CODE        TYPE      PUBLISHED  NAME\n" +
				"c1  Transistor  STANDARD  true       Transistoren\n" +
				"c2              STANDARD  false      Plain only\n",
		},
		{
			// Every level below the category; PARENT is the category each one lies directly below.
			args:    []string{"categories", "subcategories", "c/1", "--language", "nl"},
			path:    "/category/acme1/categories/c%2F1/subcategories",
			query:   "showUnpublished=true",
			scopes:  []string{"category.category_read_unpublished"},
			filters: listFilters{kinds: sortable},
			body: `[{"id":"c2","parentId":"c/1","type":"STANDARD","published":true,"localizedName":{"nl":"Zagen"}},
				{"id":"c3","parentId":"c2","type":"STANDARD","published":false,"localizedName":{"en":"Blades"}}]`,
			want: "ID  PARENT  CODE  TYPE      PUBLISHED  NAME\n" +
				"c2  c/1           STANDARD  true       Zagen\n" +
				"c3  c2            STANDARD  false      Blades (en)\n",
		},
		{
			args:    []string{"categories", "parents", "c3", "--language", "nl"},
			path:    "/category/acme1/categories/c3/parents",
			query:   "showUnpublished=true",
			scopes:  []string{"category.category_read_unpublished"},
			filters: listFilters{kinds: sortable},
			body: `[{"id":"c/1","type":"STANDARD","published":true,"localizedName":{"nl":"Werkplaats"}},
				{"id":"c2","parentId":"c/1","type":"STANDARD","published":true,"localizedName":{"nl":"Zagen"}}]`,
			want: "ID   CODE  TYPE      PUBLISHED  NAME\n" +
				"c/1        STANDARD  true       Werkplaats\n" +
				"c2         STANDARD  true       Zagen\n",
		},
		{
			// PRODUCT is the product id; Emporix sends the type in lower case.
			args:    []string{"categories", "assignments", "list", "c2", "--language", "nl"},
			path:    "/category/acme1/categories/c2/assignments",
			query:   "showUnpublished=true",
			scopes:  []string{"category.category_read_unpublished"},
			filters: listFilters{kinds: sortable},
			body: `[{"id":"a1","categoryId":"c2","ref":{"id":"100200","type":"product","published":false,
				"localizedName":{"en":"Saw","nl":"Zaag"}}}]`,
			want: "ID  PRODUCT  TYPE     PUBLISHED  NAME\na1  100200   product  false      Zaag\n",
		},
	}
}

// categoryTreeReads returns the Category Service reads of category trees.
func categoryTreeReads() []readCase {
	return []readCase{
		{
			// Each root is followed by its subcategories, depth first, indented by level; a
			// missing or null subcategories list ends a branch.
			args:   []string{"category-trees", "list", "--language", "nl"},
			path:   "/category/acme1/category-trees",
			query:  "showUnpublished=true",
			scopes: []string{"category.category_read_unpublished"},
			mode:   getWhole,
			body: `[{"id":"r1","code":"tools","published":true,"localizedName":{"en":"Tools","nl":"Werkplaats"},
				"subcategories":[{"id":"c1","published":false,"localizedName":{"en":"Saws"},"subcategories":[
					{"id":"c2","name":"Plain only","subcategories":null}]}]},
				{"id":"r2","localizedName":{"nl":"Tuin"}}]`,
			want: "ID  CODE   PUBLISHED  NAME\n" +
				"r1  tools  true       Werkplaats\n" +
				"c1         false        Saws (en)\n" +
				"c2                        Plain only\n" +
				"r2                    Tuin\n",
		},
		{
			args:   []string{"category-trees", "get", "r 1", "--language", "nl"},
			path:   "/category/acme1/category-trees/r%201",
			query:  "showUnpublished=true",
			scopes: []string{"category.category_read_unpublished"},
			mode:   getWhole,
			body: `{"id":"r 1","localizedName":{"nl":"Tuin"},
				"subcategories":[{"id":"c3","localizedName":{"nl":"Banken"}}]}`,
			want: "ID   CODE  PUBLISHED  NAME\n" +
				"r 1                   Tuin\n" +
				"c3                      Banken\n",
		},
	}
}

// catalogReads returns the Catalog and Brand Service reads.
func catalogReads() []readCase {
	return []readCase{
		{
			// The Brand Service has no tenant path segment: the tenant comes from the token.
			// A missing translation is labeled; the plain name is only a last resort.
			args:    []string{"brands", "list"},
			path:    "/brand/brands",
			filters: listFilters{kinds: filterable},
			body:    brandsBody,
			want: `ID  NAME
b1  Brand one
b2  NORTHWIND TRADE (nl)
b3  Plain only
`,
		},
		{
			args:    []string{"brands", "list", "--language", "nl"},
			path:    "/brand/brands",
			filters: listFilters{kinds: filterable},
			body:    brandsBody,
			want:    "ID  NAME\nb1  Merk een\nb2  NORTHWIND TRADE\nb3  Plain only\n",
		},
		{
			// The second catalog reproduces the live disagreement between status and visibility.
			args:    []string{"catalogs", "list"},
			path:    "/catalog/acme1/catalogs",
			filters: listFilters{kinds: sortable},
			body:    catalogsBody,
			want: `ID   NAME       STATUS   VISIBLE  FROM     TO       SITES    ROOT CATEGORIES
cc   Main       visible  true     2025-10  2028-10  NL,main  1
odd  Raar (nl)  visible  false                               0
`,
		},
		{
			args:    []string{"catalogs", "list", "--language", "nl"},
			path:    "/catalog/acme1/catalogs",
			filters: listFilters{kinds: sortable},
			body:    catalogsBody,
			want:    catalogsTableInDutch,
		},
		{
			args: []string{"catalogs", "get", "cc", "--language", "nl"},
			path: "/catalog/acme1/catalogs/cc",
			mode: getWhole,
			body: `{"id":"cc","name":{"en":"Main","nl":"Hoofd"},"status":"visible","categoryIds":["c1","c2"]}`,
			want: "ID  NAME   STATUS   VISIBLE  FROM  TO  SITES  ROOT CATEGORIES\n" +
				"cc  Hoofd  visible                            2\n",
		},
		{
			// The category id is one escaped path segment.
			args:    []string{"catalogs", "for-category", "c/1", "--language", "nl"},
			path:    "/catalog/acme1/catalogs/categories/c%2F1",
			filters: listFilters{kinds: sortable},
			body:    catalogsBody,
			want:    catalogsTableInDutch,
		},
		{
			args: []string{"brands", "get", "b1", "--language", "nl"},
			path: "/brand/brands/b1",
			mode: getWhole,
			body: `{"id":"b1","name":"Brand_1","localizedName":{"en":"Brand one","nl":"Merk een"}}`,
			want: "ID  NAME\nb1  Merk een\n",
		},
	}
}

// labelReads returns the Label Service reads.
func labelReads() []readCase {
	return []readCase{
		{
			// The Label Service has no tenant path segment. A description, often HTML, is a string or a
			// map of translations; the overlay position keeps its text.
			args:    []string{"labels", "list", "--language", "nl"},
			path:    "/label/labels",
			filters: listFilters{kinds: filterable},
			body: `[{"id":"l1","name":"new","description":{"en":"<p>New</p>","nl":"<p>Nieuw</p>"},"mediaId":"m1",
				"overlay":{"isTrue":true,"position":1.0}},{"id":"l2","name":"sale","description":"Sale"}]`,
			want: "ID  NAME  OVERLAY  POSITION  MEDIA  DESCRIPTION\n" +
				"l1  new   true     1.0       m1     <p>Nieuw</p>\n" +
				"l2  sale                            Sale\n",
		},
		{
			args: []string{"labels", "get", "l/1", "--language", "nl"},
			path: "/label/labels/l%2F1",
			mode: getWhole,
			body: `{"id":"l/1","name":"new","description":{"nl":"Nieuw"},"overlay":{"isTrue":false,"position":2}}`,
			want: "ID   NAME  OVERLAY  POSITION  MEDIA  DESCRIPTION\nl/1  new   false    2                Nieuw\n",
		},
	}
}

// mediaReads returns the Media Service reads.
func mediaReads() []readCase {
	return []readCase{
		{
			// File details are empty for a link; an asset may have no status (it is then ready);
			// the size keeps its text; REFERENCES joins type and id.
			args:    []string{"media", "list"},
			path:    "/media/acme1/assets",
			scopes:  []string{"media.asset_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"m1","type":"BLOB","access":"PRIVATE","status":"READY","details":{"filename":"a.pdf",
				"mimeType":"application/pdf","bytes":2048.0},"refIds":[{"type":"PRODUCT","id":"100200"},
				{"type":"CATEGORY","id":"c1"}]},{"id":"m2","type":"LINK","access":"PUBLIC",
				"url":"https://example.com/a.png"}]`,
			want: "ID  TYPE  ACCESS   STATUS  FILE   MIME TYPE        BYTES   REFERENCES\n" +
				"m1  BLOB  PRIVATE  READY   a.pdf  application/pdf  2048.0  PRODUCT 100200, CATEGORY c1\n" +
				"m2  LINK  PUBLIC                                           \n",
		},
		{
			// An upload that is not complete has no size yet.
			args:   []string{"media", "get", "m/1"},
			path:   "/media/acme1/assets/m%2F1",
			scopes: []string{"media.asset_read"},
			mode:   getWhole,
			body: `{"id":"m/1","type":"BLOB","access":"PRIVATE","status":"PENDING",
				"details":{"filename":"b.pdf","mimeType":"application/pdf"}}`,
			want: "ID   TYPE  ACCESS   STATUS   FILE   MIME TYPE        BYTES  REFERENCES\n" +
				"m/1  BLOB  PRIVATE  PENDING  b.pdf  application/pdf         \n",
		},
		{
			// A private file gets a signed URL that expires.
			args:   []string{"media", "download-url", "m/1"},
			path:   "/media/acme1/assets/m%2F1/download-url",
			scopes: []string{"media.asset_read"},
			mode:   getWhole,
			body:   `{"provider":"GCS","url":"https://storage.example.com/a.pdf?sig=x","expiresAt":"2026-10-14T12:15:00Z"}`,
			want: "PROVIDER  EXPIRES               URL\n" +
				"GCS       2026-10-14T12:15:00Z  https://storage.example.com/a.pdf?sig=x\n",
		},
		{
			// A link's URL does not expire.
			args:   []string{"media", "download-url", "m2"},
			path:   "/media/acme1/assets/m2/download-url",
			scopes: []string{"media.asset_read"},
			mode:   getWhole,
			body:   `{"provider":"LINK","url":"https://example.com/a.png"}`,
			want: "PROVIDER  EXPIRES  URL\n" +
				"LINK               https://example.com/a.png\n",
		},
	}
}
