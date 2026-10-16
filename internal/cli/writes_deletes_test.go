package cli

import "net/http"

// deleteWrites are the deletes of units, brands, labels, schemas and schema
// references, prices, availability, index provider configurations and
// configurations, each sent without a body unless --file names what it
// deletes, and confirmed by typing what it names, or the tenant.
func deleteWrites() []writeCase {
	return append(catalogDeleteWrites(), []writeCase{
		{
			args: []string{"units", "delete", "kg m"}, method: http.MethodDelete, path: "/unit-handling/acme1/units/kg%20m",
			scopes: []string{unitManageScope}, confirmBy: "kg m", status: http.StatusNoContent,
		},
		{
			args: []string{"units", "bulk-delete"}, file: `["g", "kg"]`, method: http.MethodDelete,
			path: "/unit-handling/acme1/units", scopes: []string{unitManageScope}, confirmBy: "acme1",
			many: "the units whose codes --file names", status: http.StatusNoContent,
		},
		{
			args: []string{"brands", "delete", "b1"}, method: http.MethodDelete, path: "/brand/brands/b1",
			scopes: []string{"brand.brand_manage"}, confirmBy: "b1", status: http.StatusNoContent,
		},
		{
			args: []string{"labels", "delete", "l1"}, method: http.MethodDelete, path: "/label/labels/l1",
			scopes: []string{"label.label_delete"}, confirmBy: "l1", status: http.StatusNoContent,
		},
		{
			args: []string{"schemas", "delete", "s1"}, method: http.MethodDelete, path: "/schema/acme1/schemas/s1",
			scopes: []string{schemaManage}, confirmBy: "s1", status: http.StatusNoContent,
		},
		{
			args: []string{"schemas", "references", "delete", "r1"}, method: http.MethodDelete,
			path: "/schema/acme1/references/r1", scopes: []string{schemaManage}, confirmBy: "r1",
			status: http.StatusNoContent,
		},
		{
			args: []string{"prices", "delete", "p1"}, method: http.MethodDelete, path: "/price/acme1/prices/p1",
			scopes: []string{priceManageScope}, confirmBy: "p1", status: http.StatusNoContent,
		},
		{
			args: []string{"availability", "delete", "100200", "--site", "AcmeStore_NL"}, method: http.MethodDelete,
			path: "/availability/acme1/availability/100200/AcmeStore_NL", scopes: []string{availabilityManageScope},
			confirmBy: "100200", status: http.StatusNoContent,
		},
		{
			args: []string{"availability", "bulk-delete"}, file: "[" + stockItem + "]", method: http.MethodDelete,
			path: "/availability/acme1/availability/bulk", scopes: []string{availabilityManageScope},
			confirmBy: "acme1", many: "the availability that --file names",
			status: http.StatusMultiStatus, body: `[{"id":"AcmeStore_NL:100200","index":0,"code":204}]`,
			want: "[\n  {\n    \"code\": 204,\n    \"id\": \"AcmeStore_NL:100200\",\n    \"index\": 0\n  }\n]\n",
		},
		{
			args: []string{"indexing", "providers", "delete", "ALGOLIA"}, method: http.MethodDelete,
			path: "/indexing/acme1/configurations/ALGOLIA", scopes: []string{indexingManageScope},
			confirmBy: "ALGOLIA", status: http.StatusNoContent,
		},
		{
			args: []string{"configurations", "delete", "project_country"}, method: http.MethodDelete,
			path: "/configuration/acme1/configurations/project_country", scopes: []string{configurationManage},
			confirmBy: "project_country", status: http.StatusNoContent,
		},
		{
			args:   []string{"configurations", "client-delete", "acme.indexing-client", "batchSize"},
			method: http.MethodDelete, path: "/configuration/acme1/clients/acme.indexing-client/configurations/batchSize",
			scopes: []string{configurationManage}, confirmBy: "batchSize", status: http.StatusNoContent,
		},
	}...)
}

// catalogDeleteWrites are the deletes of catalogs, categories, the
// assignments of products to categories, products and product templates.
func catalogDeleteWrites() []writeCase {
	categories := []string{categoryManageScope}
	return []writeCase{
		{
			args: []string{"catalogs", "delete", "c1"}, method: http.MethodDelete, path: "/catalog/acme1/catalogs/c1",
			scopes: []string{"catalog.catalog_manage"}, confirmBy: "c1", status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "delete", "c1"}, method: http.MethodDelete, path: "/category/acme1/categories/c1",
			query: "withSubcategories=false", scopes: categories, confirmBy: "c1", status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "delete", "c1", "--with-subcategories"}, method: http.MethodDelete,
			path: "/category/acme1/categories/c1", query: "withSubcategories=true", scopes: categories,
			confirmBy: "c1", many: "the category and every category below it", status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "assignments", "delete", "c1", "001 234"}, method: http.MethodDelete,
			path: "/category/acme1/categories/c1/assignments/references/001%20234", scopes: categories,
			confirmBy: "001 234", status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "assignments", "delete-assignment", "c1", "a1"}, method: http.MethodDelete,
			path: "/category/acme1/categories/c1/assignments/a1", scopes: categories, confirmBy: "a1",
			status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "assignments", "delete-all", "c1"}, method: http.MethodDelete,
			path: "/category/acme1/categories/c1/assignments", query: "showUnpublished=true",
			scopes: []string{categoryManageScope, "category.category_read_unpublished"}, confirmBy: "c1",
			many: "the assignments of every product to the category", status: http.StatusNoContent,
		},
		{
			args: []string{"categories", "assignments", "delete-product-everywhere", "100200"}, method: http.MethodDelete,
			path: "/category/acme1/assignments/references/100200", scopes: categories, confirmBy: "100200",
			many: "the assignments of the product to every category", status: http.StatusNoContent,
		},
		{
			args: []string{"products", "delete", "100200"}, method: http.MethodDelete, path: "/product/acme1/products/100200",
			query: "force=false", scopes: []string{"product.product_manage"}, confirmBy: "100200",
			status: http.StatusNoContent,
		},
		{
			args: []string{"products", "delete", "100200", "--force"}, method: http.MethodDelete,
			path: "/product/acme1/products/100200", query: "force=true", scopes: []string{"product.product_manage"},
			confirmBy: "100200", many: "the product and every variant below it", status: http.StatusNoContent,
		},
		{
			args: []string{"product-templates", "delete", "t1"}, method: http.MethodDelete,
			path: "/product/acme1/product-templates/t1", scopes: []string{"product.product_template_manage"},
			confirmBy: "t1", status: http.StatusNoContent,
		},
	}
}
