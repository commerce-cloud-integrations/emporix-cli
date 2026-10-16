package cli

import (
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

// productManageScope is the scope of every product write, and of the reads
// of the variant recalculation jobs (opt-in).
const productManageScope = "product.product_manage"

// maxProductsPerRequest is the most products that a bulk write of the
// Product Service takes in one request.
const maxProductsPerRequest = 200

// productPublishing returns the choice of a write of products sent with
// method, which only adds scopes: those that Emporix requires besides
// product.product_manage only to publish a product, which every write of
// products may do, and to unpublish one, which a write that replaces or
// changes products (not POST) may do, as the reference names them. Its flags
// may be given together. Emporix refuses a product that the write publishes
// or unpublishes without the flag that allows it.
func productPublishing(method string) queryChoice {
	options := []choiceOption{{flag: "allow-publish", scopes: []string{"product.product_publish"},
		usage:   "allow the write to publish products, which needs a scope of its own",
		refused: "a product that the write publishes"}}
	if method != http.MethodPost {
		options = append(options, choiceOption{flag: "allow-unpublish", scopes: []string{"product.product_unpublish"},
			usage:   "allow the write to unpublish products, which needs a scope of its own",
			refused: "a product that the write unpublishes"})
	}
	return queryChoice{options: options}
}

// productQuery returns the query parameters that every write of products
// takes: doIndex, sent when its flag is given, and the switches.
func productQuery() (queryFlags, []paramFlag) {
	return queryFlags{optional: []paramFlag{{
		param: "doIndex", flag: "do-index", choices: []string{queryTrue, queryFalse},
		usage: "whether Emporix reindexes the products it writes (by default, it does; with false, other changes " +
			"or Emporix's scheduled indexing may still index them, as it may index a product published with false)",
	}}}, []paramFlag{{
		param: "skipVariantGeneration", flag: "skip-variant-generation",
		usage: "do not generate the variants of a PARENT_VARIANT product from its variant attributes",
	}, {
		param: "skipRelatedItemsValidation", flag: "skip-related-items-validation",
		usage: "do not check that related products exist, as when they are written later (a tenant may " +
			"ignore it and refuse unknown related products all the same)",
	}}
}

// productConflict says when Emporix refuses a write of a product as a conflict.
const productConflict = "Emporix refuses it as a conflict (exit 5) when another product has its code or, given\n" +
	"metadata.version, when the product changed since that version."

// productBulkAnswer says how Emporix answers a bulk write of products.
const productBulkAnswer = "Emporix answers each product with its index, id, code (the HTTP status of the " +
	"product, e.g.\n201) and status (e.g. CREATED), and a message where it failed."

// productWrites returns the writes of products: create, update ID, patch ID,
// bulk-create, bulk-update, import and recalculate.
func productWrites(a *app, products resource[product]) []*cobra.Command {
	scopes := []string{productManageScope}
	most := strconv.Itoa(maxProductsPerRequest)
	declared := &bulkWrites{
		path: products.path, nouns: products.nouns(), scopes: scopes, maxItems: maxProductsPerRequest,
		create: "Creates the products in --file, a JSON array of products, each as create takes it.\n" +
			productBulkAnswer,
		update: "Upserts the products in --file, a JSON array of products, each complete and with its id, as\n" +
			"update takes it. import sends more than " + most + " in batches.\n" + productBulkAnswer,
	}
	bulk := declared.writes()
	batches := bulk[1]
	batches.use, batches.short = importUse, "Upsert the products of a file in batches"
	batches.long = "Upserts the products in --file, a JSON array of any number of products, each complete and\n" +
		"with its id as update takes it, as bulk-update does up to " + most + ". To resume an import that\n" +
		"stopped, send the products from the item that its error names: a product sent again is\n" +
		"upserted again, but one whose metadata.version the import changed is a conflict.\n" + productBulkAnswer
	batches.body = &payload{maxItems: maxProductsPerRequest, batched: true}
	writes := append([]write{{
		use: createUse, short: "Create a product", method: http.MethodPost, req: request{path: products.path},
		body: &payload{},
		long: "Creates the product in the JSON object of --file, as the Product Service takes it (without an\n" +
			"id, Emporix makes one), and prints its id. Emporix refuses it as a conflict (exit 5) when a\n" +
			"product has its id or code.",
	}, {
		use: products.itemUse("update"), short: "Upsert a product", method: http.MethodPut,
		req:      request{path: products.itemPath()},
		reserved: declared.reserved(http.MethodPut, products.itemPath()),
		ids:      products.itemIDs(),
		body:     &payload{},
		long: "Replaces product ID with the JSON object of --file, the complete product: an upsert, as\n" +
			"Emporix creates the product where none has the ID (and answers with its id and yrn). Mixins\n" +
			"left out are removed.\n" + productConflict,
	}, {
		use: products.itemUse("patch"), short: "Change members of a product", method: http.MethodPatch,
		req: request{path: products.itemPath()}, body: &payload{}, ids: products.itemIDs(),
		long: "Changes the members of product ID that the JSON object of --file names; the others stay,\n" +
			"and mixins are merged member by member (update removes a mixin).\n" + productConflict,
	}}, append(bulk, batches)...)
	params, switches := productQuery()
	for i := range writes {
		writes[i].scopes, writes[i].language = scopes, stringsOrMaps
		writes[i].choice = productPublishing(writes[i].method)
		writes[i].params, writes[i].switches = params, switches
	}
	return append(writeCommands(a, writes), (&write{
		use: recalculate, short: "Recalculate the variant trees of dynamic variant products",
		method: http.MethodPost, req: request{path: products.path + "/recalculate"}, scopes: scopes,
		long: "Starts the recalculation of the variant trees of DYNAMIC_VARIANT products once their products\n" +
			"are written in bulk (a write of one product updates its tree at once). Emporix finds the root\n" +
			"product of each ID, at any level, and starts one job per root, which\n" +
			"`epx products recalculation-jobs list` shows (jobs). An ID that is unknown, is not of a\n" +
			"DYNAMIC_VARIANT product, or whose root has a job pending or processing is skipped\n" +
			"(skippedProductIds, left out when none is); where every ID is, the answer has no jobs.\n" +
			"Takes up to 1000 IDs, given with --product or in the JSON object of --file\n" +
			"({\"productIds\": [...]}).",
		body: &payload{typed: objectBody{
			listField("productIds", "product", "`ID` of a product to recalculate (repeatable)").require(),
		}},
	}).command(a))
}

// productTemplateWrites returns the writes of product templates: create and
// update ID.
func productTemplateWrites(a *app, templates resource[productTemplate]) []*cobra.Command {
	scopes := []string{"product.product_template_manage"}
	return writeCommands(a, []write{{
		use: createUse, short: "Create a product template", method: http.MethodPost,
		req: request{path: templates.path}, scopes: scopes, body: &payload{}, language: stringsOrMaps,
		long: "Creates the product template in the JSON object of --file (name, attributes) and prints its id.",
	}, {
		use: templates.itemUse("update"), short: "Replace a product template", method: http.MethodPut,
		req: request{path: templates.itemPath()}, scopes: scopes, body: &payload{}, language: stringsOrMaps,
		ids: templates.itemIDs(),
		long: "Replaces product template ID with the JSON object of --file, which names its\n" +
			"metadata.version: Emporix refuses it as a conflict (exit 5) when the template changed since.\n" +
			"Every update makes a new version of the template; the products based on an earlier version keep\n" +
			"it until they are updated to the new one. An unknown ID is not found (exit 4).",
	}})
}
