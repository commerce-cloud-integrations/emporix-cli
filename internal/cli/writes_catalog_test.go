package cli

import "net/http"

// brandAndLabelWrites are the writes of brands and labels, whose services
// take the tenant from the token: create, update ID and patch ID each.
func brandAndLabelWrites() []writeCase {
	brands, labels := []string{"brand.brand_manage"}, []string{"label.label_manage"}
	return []writeCase{
		{
			args: []string{"brands", "create"}, file: `{"name":"Acme","localizedName":{"en":"Acme","de":"Acme"}}`,
			method: http.MethodPost, path: "/brand/brands", scopes: brands, lang: "*", status: http.StatusCreated,
			body: `{"id":"b1","name":"Acme","metadata":{"version":1}}`,
			want: "{\n  \"id\": \"b1\",\n  \"metadata\": {\n    \"version\": 1\n  },\n  \"name\": \"Acme\"\n}\n",
		},
		{
			// The ID is escaped as a path segment.
			args: []string{"brands", "update", "b 1"}, file: `{"name":"Acme","metadata":{"version":1}}`,
			method: http.MethodPut, path: "/brand/brands/b%201", scopes: brands, lang: "*", status: http.StatusOK,
			body: `{"id":"b 1","name":"Acme"}`, want: "{\n  \"id\": \"b 1\",\n  \"name\": \"Acme\"\n}\n",
		},
		{
			args: []string{"brands", "patch", "b1"}, file: `{"description":"<p>Acme</p>"}`, method: http.MethodPatch,
			path: "/brand/brands/b1", scopes: brands, lang: "*", status: http.StatusOK, body: `{"id":"b1"}`,
			want: "{\n  \"id\": \"b1\"\n}\n",
		},
		{
			args:   []string{"labels", "create"},
			file:   `{"name":"Free shipping","description":"<p>Ships free</p>","overlay":{"isTrue":true,"position":1}}`,
			method: http.MethodPost, path: "/label/labels", scopes: labels, status: http.StatusCreated,
			body: `{"id":"l1","name":"Free shipping"}`, want: "{\n  \"id\": \"l1\",\n  \"name\": \"Free shipping\"\n}\n",
		},
		{
			args: []string{"labels", "update", "l1"}, file: `{"name":"Free shipping","description":"Ships free"}`,
			method: http.MethodPut, path: "/label/labels/l1", scopes: labels, status: http.StatusOK, body: `{"id":"l1"}`,
			want: "{\n  \"id\": \"l1\"\n}\n",
		},
		{
			// Emporix takes a PATCH of only the overlay, though the reference requires name and description.
			args: []string{"labels", "patch", "l1"}, file: `{"overlay":{"position":0}}`,
			method: http.MethodPatch, path: "/label/labels/l1", scopes: labels, status: http.StatusNoContent,
		},
	}
}

// The write cases of the Catalog and Category services (see TestWrites).

// assignmentStatuses is the answer of a bulk write of category assignments in
// which both items succeeded (207 Multi-Status), and assignmentStatusesJSON
// what epx prints of it.
const (
	assignmentStatuses = `[{"assignmentId":"a1","productId":"100200","index":0,"code":201,"status":"CREATED"},
		{"assignmentId":"a2","productId":"100201","index":1,"code":200,"status":"OK"}]`
	assignmentStatusesJSON = "[\n  {\n    \"assignmentId\": \"a1\",\n    \"code\": 201,\n    \"index\": 0,\n" +
		"    \"productId\": \"100200\",\n    \"status\": \"CREATED\"\n  },\n  {\n    \"assignmentId\": \"a2\",\n" +
		"    \"code\": 200,\n    \"index\": 1,\n    \"productId\": \"100201\",\n    \"status\": \"OK\"\n  }\n]\n"
	// assignmentsBody is the body of a bulk write of category assignments, and
	// assignmentRefsBody that of their upsert, which names no type.
	assignmentsBody    = `[{"ref":{"id":"100200","type":"PRODUCT"}},{"ref":{"id":"100201","type":"PRODUCT"}}]`
	assignmentRefsBody = `[{"ref":{"id":"100200"}},{"ref":{"id":"100201"}}]`
)

// catalogWriteCases returns the writes of catalogs, categories, their
// assignments and category trees.
func catalogWriteCases() []writeCase {
	manage := []string{categoryManageScope}
	return []writeCase{
		{
			args: []string{"catalogs", "create"}, file: `{"id":"AcmeStore_catalog","name":{"en":"Acme Store"}}`,
			method: http.MethodPost, path: "/catalog/acme1/catalogs", scopes: []string{"catalog.catalog_manage"},
			status: http.StatusCreated, body: `{"id":"AcmeStore_catalog"}`, want: "{\n  \"id\": \"AcmeStore_catalog\"\n}\n",
		},
		{
			args: []string{"catalogs", "update", "AcmeStore_catalog"}, file: `{"name":{"en":"Acme"},"metadata":{"version":2}}`,
			method: http.MethodPut, path: "/catalog/acme1/catalogs/AcmeStore_catalog",
			scopes: []string{"catalog.catalog_manage"}, status: http.StatusNoContent,
		},
		{
			// The id is one escaped path segment.
			args: []string{"catalogs", "patch", "Acme Store"}, file: `{"publishedSites":["AcmeStore_NL"]}`,
			method: http.MethodPatch, path: "/catalog/acme1/catalogs/Acme%20Store",
			scopes: []string{"catalog.catalog_manage"}, status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "create"}, file: `{"localizedName":{"en":"Saws"},"published":false}`,
			method: http.MethodPost, path: "/category/acme1/categories", scopes: manage, status: http.StatusCreated,
			body: `{"id":"c1"}`, want: "{\n  \"id\": \"c1\"\n}\n",
		},
		{
			// Publishing sends publish=true and requests the scope that Emporix requires for it.
			args: []string{"categories", "create", "--publish"}, file: `{"localizedName":{"en":"Saws"},"published":true}`,
			method: http.MethodPost, path: "/category/acme1/categories", query: "publish=true",
			scopes: []string{categoryManageScope, "category.category_publish"}, status: http.StatusCreated,
			body: `{"id":"c1"}`, want: "{\n  \"id\": \"c1\"\n}\n",
		},
		{
			args: []string{"categories", "update", "c1"}, file: `{"localizedName":{"en":"Saws"},"published":false}`,
			method: http.MethodPut, path: "/category/acme1/categories/c1", scopes: manage, status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "update", "c1", "--unpublish"}, file: `{"published":false,"metadata":{"version":4}}`,
			method: http.MethodPut, path: "/category/acme1/categories/c1", query: "publish=false",
			scopes: []string{categoryManageScope, "category.category_unpublish"}, status: http.StatusNoContent,
		},
		{
			// Emporix publishes it as the body says, without the publish parameter or scope (live).
			args: []string{"categories", "patch", "c1"}, file: `{"published":true,"metadata":{"version":4}}`,
			method: http.MethodPatch, path: "/category/acme1/categories/c1", scopes: manage,
			status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "assignments", "create", "c1"}, file: `{"ref":{"id":"100200","type":"PRODUCT"}}`,
			method: http.MethodPost, path: "/category/acme1/categories/c1/assignments", scopes: manage,
			status: http.StatusCreated, body: `{"id":"a1"}`, want: "{\n  \"id\": \"a1\"\n}\n",
		},
		{
			args: []string{"categories", "assignments", "bulk-create", "c1"}, file: assignmentsBody,
			method: http.MethodPost, path: "/category/acme1/categories/c1/assignments/bulk", scopes: manage,
			status: http.StatusMultiStatus, body: assignmentStatuses, want: assignmentStatusesJSON,
		},
		{
			args: []string{"categories", "assignments", "bulk-update", "c1"}, file: assignmentRefsBody,
			method: http.MethodPut, path: "/category/acme1/categories/c1/assignments/references/bulk", scopes: manage,
			status: http.StatusMultiStatus, body: assignmentStatuses, want: assignmentStatusesJSON,
		},
		{
			// No body; the product id is one escaped path segment.
			args:   []string{"categories", "assignments", "update", "c1", "001 234"},
			method: http.MethodPut, path: "/category/acme1/categories/c1/assignments/references/001%20234",
			scopes: manage, status: http.StatusCreated, body: `{"id":"a1"}`, want: "{\n  \"id\": \"a1\"\n}\n",
		},
		{
			// No body; the tree is printed whole, unpublished categories too.
			args:   []string{"category-trees", "rebuild", "r1"},
			method: http.MethodPost, path: "/category/acme1/category-trees/r1/rebuild", query: "showUnpublished=true",
			scopes: []string{categoryManageScope, "category.category_read_unpublished"}, status: http.StatusOK,
			body: `{"id":"r1","subcategories":[{"id":"c1","published":false}]}`,
			want: "{\n  \"id\": \"r1\",\n  \"subcategories\": [\n    {\n      \"id\": \"c1\",\n" +
				"      \"published\": false\n    }\n  ]\n}\n",
		},
	}
}
