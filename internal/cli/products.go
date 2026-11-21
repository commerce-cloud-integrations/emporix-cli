package cli

import (
	"slices"
	"strconv"

	"github.com/spf13/cobra"
)

type product struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	ProductType string    `json:"productType"`
	Published   *bool     `json:"published"`
	Name        localized `json:"name"`
}

const (
	productID     = "PRODUCT_ID"
	forProductUse = "for-product " + productID
	productIDNote = productID + " is the product id (as in `epx products list`), not its code."
)

const productsUse = "products"

const productReadUnpublishedScope = "product.product_read_unpublished"

func newProductsCommand(a *app) *cobra.Command {
	r := products()
	del := itemDelete(r, "Delete a product",
		"Deletes the product ID. A parent variant product (PARENT_VARIANT, or DYNAMIC_VARIANT with "+
			"variants below it) is deleted only with --force, with every variant below it "+
			"(without it, Emporix answers 400 Bad Request, exit 1). An unknown ID is no error: Emporix "+
			"answers as for a deleted product.",
		[]string{"product.product_manage"})
	del.switches = forceSwitch("delete a parent variant product with every variant below it")
	del.deletes = deletion{many: "the product and every variant below it", with: forceParam}
	reads := []*cobra.Command{r.searchCommand(a,
		"Reads the products, published and unpublished, that an Emporix q filter finds (--query; "+
			"without it, every product).", queryField()), newRecalculationJobsCommand(a)}
	return newResourceCommand(a, r, slices.Concat(reads, productWrites(a, r), []*cobra.Command{del.command(a)})...)
}

func products() resource[product] {
	return resource[product]{
		use:   productsUse,
		short: "Read, create, update and delete products",
		about: "Reads and changes products. Reads find products published and unpublished (without the " +
			productReadUnpublishedScope + " scope, Emporix would silently omit unpublished products). " +
			"create, update and patch change one product, bulk-create and bulk-update up to " +
			strconv.Itoa(maxProductsPerRequest) + ", import any number in batches, each from --file; " +
			"recalculate rebuilds variant trees; delete deletes one.",
		long: "Reads products, published and unpublished (without the " + productReadUnpublishedScope + " " +
			"scope, Emporix would silently omit unpublished products).",
		path:    "/product/{tenant}/products",
		filters: listFilters{kinds: filterable},
		scopes:  []string{productReadUnpublishedScope},
		view: view[product]{
			singular:  "product",
			header:    []string{columnID, columnCode, columnType, columnPublished, columnName},
			localized: true,
			siteAware: true,
			row: func(p product, lang language) []string {
				return []string{p.ID, p.Code, p.ProductType, formatBool(p.Published), p.Name.in(lang)}
			},
		},
	}
}
