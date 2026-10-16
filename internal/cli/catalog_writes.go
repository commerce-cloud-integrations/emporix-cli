package cli

import (
	"net/http"
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// The writes of the Catalog and Category services: catalogs, categories, the
// products assigned to a category, and the rebuild of a category tree. Their
// localized members are maps of languages to text whatever the
// Content-Language, so they take no --content-language.

// versionNote says how a replacement (PUT) of one item avoids overwriting a
// change made since the item was read, the item being a singular that getLine
// reads.
func versionNote(singular, getLine string) string {
	return "With metadata.version, as `epx " + getLine + " -o json` shows it, Emporix refuses\n" +
		"the write as a conflict (exit 5) when the " + singular + " changed since it was read."
}

// catalogWrites returns the writes of catalogs: create, update ID (an upsert),
// patch ID and delete ID.
func catalogWrites(a *app, catalogs resource[catalog]) []*cobra.Command {
	manage := []string{"catalog.catalog_manage"}
	version := versionNote(catalogs.singular, "catalogs get ID")
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a catalog", method: http.MethodPost,
			long: "Creates a catalog from the JSON in --file, shaped as `epx catalogs get ID -o json` shows\n" +
				"one: name and description (maps of languages to text), visibility, publishedSites and\n" +
				"categoryIds (root categories); Emporix generates the id unless the body gives one. It\n" +
				"answers with the id; an id that exists already is a conflict (exit 5).",
			req: request{path: catalogs.path}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: catalogs.itemUse("update"), short: "Upsert a catalog", method: http.MethodPut,
			long: "Replaces the catalog ID with the JSON in --file, or creates it with that id when there\n" +
				"is none (an upsert); the body has the members of create's but the id.\n" + version,
			req: request{path: catalogs.itemPath()}, scopes: manage, body: &payload{}, ids: catalogs.itemIDs(),
		}).command(a),
		(&write{
			use: catalogs.itemUse("patch"), short: "Change members of a catalog", method: http.MethodPatch,
			long: "Changes the members of the catalog ID that the JSON in --file gives, and keeps the\n" +
				"others. Emporix requires metadata.version in the body, as `epx catalogs get ID -o json`\n" +
				"shows it, and answers a version that is not the catalog's as it answers an unknown\n" +
				"catalog, with 404 Not Found: epx then exits with 5, a conflict (read the catalog again\n" +
				"before you retry), also for an unknown ID, and with 4 without a version.",
			req: request{path: catalogs.itemPath()}, scopes: manage, body: &payload{}, ids: catalogs.itemIDs(),
			answers: emporix.Answers{StaleNotFound: true},
		}).command(a),
		itemDelete(catalogs, "Delete a catalog", "Deletes the catalog ID, as `epx catalogs list` shows it.",
			manage).command(a),
	}
}

// categoryManageScope is the scope of every write of categories and their
// assignments.
const categoryManageScope = "category.category_manage"

// publishing is the choice of a write of a category to publish it, with
// --publish, or also to unpublish it, with --unpublish: the publish query
// parameter, which Emporix requires to match published in the body, each
// value with the scope it requires for it.
func publishing(unpublish bool) queryChoice {
	options := []choiceOption{{
		flag: "publish", value: queryTrue, scopes: []string{"category.category_publish"},
		usage: "publish the category: sends publish=true, which needs published: true in the body",
	}}
	if unpublish {
		options = append(options, choiceOption{
			flag: "unpublish", value: queryFalse, scopes: []string{"category.category_unpublish"},
			usage: "unpublish the category: sends publish=false, which needs published: false in the body",
		})
	}
	return queryChoice{param: "publish", options: options}
}

// categoryWrites returns the writes of categories: create, update ID (an
// upsert), patch ID and delete ID. create and update publish or unpublish a
// category only with --publish or --unpublish (publishing); patch applies
// published from its body as sent, and Emporix ignores the publish parameter
// there (live), so it takes neither. A root category with categories below it
// is deleted only with --with-subcategories.
func categoryWrites(a *app, categories resource[category]) []*cobra.Command {
	manage := []string{categoryManageScope}
	del := itemDelete(categories, "Delete a category",
		"Deletes the category ID. Without --with-subcategories, the categories directly below it\n"+
			"move to the category above it and keep the categories below them; Emporix refuses to\n"+
			"delete a root category with categories below it (400 Bad Request, exit 1). With it,\n"+
			"Emporix deletes every category below it too.", manage)
	del.switches = []paramFlag{{param: "withSubcategories", flag: "with-subcategories",
		usage: "delete every category below the category too"}}
	del.deletes = deletion{many: "the category and every category below it", with: "with-subcategories"}
	version := versionNote(categories.singular, "categories get ID")
	const published = "To publish or unpublish it, give published in the body and --publish or --unpublish,\n" +
		"which sends the publish parameter that Emporix requires to match it, with the scope it\n" +
		"requires for it. Emporix requires --publish for any body with published: true, also of a\n" +
		"category that is published already."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a category", method: http.MethodPost,
			long: "Creates a category from the JSON in --file, shaped as `epx categories get ID -o json`\n" +
				"shows one: localizedName (a map of languages to text, as are the other localized\n" +
				"members), published, which Emporix requires, parentId (a root category without one)\n" +
				"and the other members; Emporix generates the id unless the body gives one, and answers\n" +
				"with it. To create it published, give published: true in the body and --publish, which\n" +
				"sends the publish parameter that Emporix requires to match it, with the scope it requires\n" +
				"for it.",
			req: request{path: categories.path}, scopes: manage, body: &payload{}, choice: publishing(false),
		}).command(a),
		(&write{
			use: categories.itemUse("update"), short: "Upsert a category", method: http.MethodPut,
			long: "Replaces the category ID with the JSON in --file, or creates it with that id when\n" +
				"there is none (an upsert); the body has the members of create's, published included, but\n" +
				"id, type and ownClassificationMixins. Without parentId, the category becomes a root\n" +
				"category.\n" +
				version + "\n" + published,
			req: request{path: categories.itemPath()}, scopes: manage, body: &payload{}, ids: categories.itemIDs(),
			choice: publishing(true),
		}).command(a),
		(&write{
			use: categories.itemUse("patch"), short: "Change members of a category", method: http.MethodPatch,
			long: "Changes the members of the category ID that the JSON in --file gives, and keeps the\n" +
				"others. Emporix requires metadata.version in the body, as `epx categories get ID -o json`\n" +
				"shows it, and refuses a version that is not the category's with 400 Bad Request (exit 1),\n" +
				"not as a conflict. published in the body publishes or unpublishes the category as sent,\n" +
				"without the publish scopes. A move (parentId) shows in `epx categories parents` at once,\n" +
				"in `epx category-trees` a few seconds later.",
			req: request{path: categories.itemPath()}, scopes: manage, body: &payload{}, ids: categories.itemIDs(),
		}).command(a),
		del.command(a),
	}
}

// maxAssignments is the most assignments that a bulk write of a category's
// assignments sends, as its reference documents.
const maxAssignments = 200

// referencesSegment follows the path of a category's assignments in the path
// of the assignment of one product, which the product's id ends, and of
// bulk-update.
const referencesSegment = "/references/"

// assignmentWrites returns the writes of the products assigned to a category
// of categories, assignments being their list: create, bulk-create and
// bulk-update CATEGORY_ID, and update CATEGORY_ID PRODUCT_ID. Each names the
// category first and says where its id comes from, as the list does. The
// deletes follow (assignmentDeletes).
func assignmentWrites(a *app, categories resource[category], assignments child[categoryAssignment]) []*cobra.Command {
	path, category := assignments.below(categories.path), categories.argument()
	ofCategory := argumentIDs{category: categories.ids()}
	about := "\n" + categories.argumentSentence(notFound)
	manage := []string{categoryManageScope}
	const ref = `{"ref": {"id": "100200", "type": "PRODUCT"}}` // an assignment that create and bulk-create send
	return []*cobra.Command{
		(&write{
			use: createUse + " " + category, short: "Assign a product to a category", method: http.MethodPost,
			long: "Assigns a product to a category: the JSON in --file names it by its id, as in\n" +
				ref + ". Emporix answers with the id of the assignment.\n" +
				"`epx categories assignments update " + category + " PRODUCT_ID` assigns one without a body." + about,
			req: request{path: path}, scopes: manage, body: &payload{}, ids: ofCategory,
		}).command(a),
		(&write{
			use: bulkCreateUse + " " + category, short: "Assign products to a category", method: http.MethodPost,
			long: "Assigns products to a category: the JSON in --file is an array of assignments, each as\n" +
				"in " + ref + "." + about,
			req: request{path: path + "/" + bulkSegment}, scopes: manage, body: &payload{maxItems: maxAssignments},
			ids: ofCategory,
		}).command(a),
		(&write{
			use: bulkUpdateUse + " " + category, short: "Upsert the assignments of products to a category",
			method: http.MethodPut,
			long: "Assigns products to a category, or updates the assignments that exist (an upsert): the\n" +
				"JSON in --file is an array of assignments, each naming a product by its id, as in\n" +
				`{"ref": {"id": "100200"}}.` + about,
			req: request{path: path + referencesSegment + bulkSegment}, scopes: manage,
			body: &payload{maxItems: maxAssignments}, ids: ofCategory,
		}).command(a),
		(&write{
			use: "update " + category + " " + productID, short: "Upsert the assignment of a product to a category",
			method: http.MethodPut,
			long: "Assigns the product PRODUCT_ID to a category, or updates its assignment when it exists (an\n" +
				"upsert). It sends no body." + productIDNote + about,
			req: request{path: path + referencesSegment + placeholderID}, scopes: manage, body: noBody{},
			reserved: map[string]string{bulkSegment: bulkUpdateUse},
			ids:      argumentIDs{category: categories.ids(), productID: products().ids()},
		}).command(a),
	}
}

// assignmentDeletes returns the deletes of the products assigned to a
// category of categories, below path, the list of one category's
// assignments: delete CATEGORY_ID PRODUCT_ID, the inverse of update, and
// delete-assignment CATEGORY_ID ASSIGNMENT_ID, which delete one, delete-all
// CATEGORY_ID, which deletes those of every product of a category, and
// delete-product-everywhere PRODUCT_ID, which deletes those of a product in
// every category. Each completes the category from categories and the
// product or assignment from the category's assignments.
func assignmentDeletes(a *app, categories resource[category], assignments child[categoryAssignment]) []*cobra.Command {
	path, manage := assignments.below(categories.path), []string{categoryManageScope}
	category := categories.argument()
	ofCategory := categories.ids()
	list := "`epx categories assignments list " + category + "`"
	const stay = " The product and the category stay."
	// delete-all finds an unpublished category only with the query and scopes of
	// the list (live), which the other deletes do not need.
	all := deleteWrite("delete-all "+category, "Unassign every product from a category",
		"Deletes the assignments of every product to the category "+category+", published or\n"+
			"unpublished, as "+list+" lists them. The products and the category stay.",
		request{path: path, query: assignments.query}, slices.Concat(manage, categories.scopes))
	all.deletes = deletion{many: "the assignments of every product to the category"}
	all.ids = argumentIDs{category: ofCategory}
	everywhere := deleteWrite("delete-product-everywhere PRODUCT_ID", "Unassign a product from every category",
		"Deletes the assignments of the product PRODUCT_ID to every category, as\n"+
			"`epx categories for-product PRODUCT_ID` lists them. The product and the categories stay."+productIDNote,
		request{path: "/category/{tenant}/assignments/references/" + placeholderID}, manage)
	everywhere.deletes = deletion{many: "the assignments of the product to every category"}
	everywhere.ids = argumentIDs{productID: products().ids()}
	one := deleteWrite(deleteUse+" "+category+" "+productID, "Unassign a product from a category",
		"Deletes the assignment of the product PRODUCT_ID to the category "+category+": the PRODUCT\n"+
			"that "+list+" shows, as update assigns it."+stay+productIDNote,
		request{path: path + referencesSegment + placeholderID}, manage)
	one.ids = argumentIDs{category: ofCategory, productID: childIDs(categories, assignments, columnProduct)}
	byID := deleteWrite("delete-assignment "+category+" ASSIGNMENT_ID", "Delete an assignment of a product to a category",
		"Deletes the assignment ASSIGNMENT_ID of a product to the category "+category+": the ID that\n"+
			list+" shows."+stay, request{path: path + "/" + placeholderID}, manage)
	byID.ids = argumentIDs{category: ofCategory, "ASSIGNMENT_ID": childIDs(categories, assignments, columnID)}
	return []*cobra.Command{one.command(a), byID.command(a), all.command(a), everywhere.command(a)}
}

// rebuildCommand returns `rebuild ID`, which rebuilds the category tree of
// trees, a root category, and prints it whole: published and unpublished
// categories, as the reads of category trees show them.
func rebuildCommand(a *app, trees resource[treeCategory]) *cobra.Command {
	return (&write{
		use:   trees.itemUse("rebuild"),
		short: "Rebuild a category tree",
		long: "Rebuilds the category tree of the root category ID (an unindented row of\n" +
			"`epx " + trees.listLine() + "`) and prints the rebuilt tree as JSON, published and unpublished\n" +
			"categories. It sends no body; another category's id is not found (exit 4).",
		method: http.MethodPost, req: request{path: trees.itemPath() + "/rebuild", query: trees.query},
		scopes: slices.Concat([]string{categoryManageScope}, trees.scopes), body: noBody{}, ids: trees.itemIDs(),
	}).command(a)
}
