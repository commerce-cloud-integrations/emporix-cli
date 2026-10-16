package cli

import (
	"slices"
	"strconv"

	"github.com/spf13/cobra"
)

// product is the part of a Product Service product shown in tables.
type product struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	ProductType string    `json:"productType"`
	Published   *bool     `json:"published"`
	Name        localized `json:"name"`
}

// The argument that names one product, the reads of what applies to it
// (their name) and the note that follows the help of a command that takes
// it, whose IDs complete with those of `epx products list` (products).
const (
	productID     = "PRODUCT_ID"
	forProductUse = "for-product " + productID
	productIDNote = "\n" + productID + " is the product id (as in `epx products list`), not its code."
)

// productsUse is the name of the group of products, its reads and writes.
const productsUse = "products"

// productReadUnpublishedScope is the scope without which Emporix silently
// omits unpublished products.
const productReadUnpublishedScope = "product.product_read_unpublished"

// newProductsCommand reads products, including unpublished ones: without the
// product.product_read_unpublished scope Emporix silently omits them, so the
// scope is always requested and a key without it fails with exit code 3.
// search finds them with a q filter sent in the request body. The group also
// reads the variant recalculation jobs (opt-in) and holds the product writes,
// deletes included.
func newProductsCommand(a *app) *cobra.Command {
	r := products()
	del := itemDelete(r, "Delete a product",
		"Deletes the product ID. A parent variant product (PARENT_VARIANT, or DYNAMIC_VARIANT with\n"+
			"variants below it) is deleted only with --force, with every variant below it\n"+
			"(without it, Emporix answers 400 Bad Request, exit 1). An unknown ID is no error: Emporix\n"+
			"answers as for a deleted product.",
		[]string{"product.product_manage"})
	del.switches = forceSwitch("delete a parent variant product with every variant below it")
	del.deletes = deletion{many: "the product and every variant below it", with: forceParam}
	reads := []*cobra.Command{r.searchCommand(a,
		"Reads the products, published and unpublished, that an Emporix q filter finds (--query;\n"+
			"without it, every product).", queryField()), newRecalculationJobsCommand(a)}
	return newResourceCommand(a, r, slices.Concat(reads, productWrites(a, r), []*cobra.Command{del.command(a)})...)
}

// products declares the products, whose IDs every PRODUCT_ID
// completes with (forProductUse).
func products() resource[product] {
	return resource[product]{
		use:   productsUse,
		short: "Read, create, update and delete products",
		about: "Reads and changes products. Reads find products published and unpublished (without the\n" +
			productReadUnpublishedScope + " scope, Emporix would silently omit unpublished products).\n" +
			"create, update and patch change one product, bulk-create and bulk-update up to " +
			strconv.Itoa(maxProductsPerRequest) + ",\nimport any number in batches, each from --file; " +
			"recalculate rebuilds variant trees; delete deletes one.",
		long: "Reads products, published and unpublished (without the " + productReadUnpublishedScope + "\n" +
			"scope, Emporix would silently omit unpublished products).",
		path:    "/product/{tenant}/products",
		filters: listFilters{kinds: filterable},
		scopes:  []string{productReadUnpublishedScope},
		view: view[product]{
			singular:  "product",
			header:    []string{columnID, columnCode, columnType, columnPublished, columnName},
			localized: true,
			row: func(p product, lang language) []string {
				return []string{p.ID, p.Code, p.ProductType, formatBool(p.Published), p.Name.in(lang)}
			},
		},
	}
}
