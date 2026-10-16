package cli

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// treeCategory is a category of a category tree with its level in the tree:
// 0 for a root category.
type treeCategory struct {
	Level    int      `json:"level"`
	Category category `json:"category"`
}

// categoryTrees declares the category trees, whose root categories
// the CATEGORY_ID of catalogs for-category completes with too.
func categoryTrees() resource[treeCategory] {
	return resource[treeCategory]{
		use:   "category-trees",
		short: "Read and rebuild category trees",
		long: "Reads category trees, published and unpublished: each root category followed by its\n" +
			"subcategories, depth first, in the order Emporix sends them; NAME is indented by the\n" +
			"category's level. get takes the id of a root category: Emporix answers another\n" +
			"category's id as not found. list's --limit counts trees, each shown whole. A move of a\n" +
			"category shows in its tree a few seconds later; a root category moved below another may\n" +
			"keep its old tree until it is deleted.",
		path:   "/category/{tenant}/category-trees",
		query:  withUnpublished(),
		scopes: []string{"category.category_read_unpublished"},
		read:   (*emporix.Client).GetUnpaged,
		items:  treeItems,
		item:   oneTreeItems,
		roots:  treeRoots, // get takes a root category only
		view: view[treeCategory]{
			singular:  "category tree",
			header:    []string{columnID, columnCode, columnPublished, columnName},
			localized: true,
			row: func(c treeCategory, lang language) []string {
				// Only the indentation shows the level: a name's own leading spaces are dropped.
				name := strings.TrimLeftFunc(output.Sanitize(c.Category.displayName(lang)), unicode.IsSpace)
				return []string{
					c.Category.ID, c.Category.Code, formatBool(c.Category.Published), strings.Repeat("  ", c.Level) + name,
				}
			},
		},
	}
}

// newCategoryTreesCommand reads category trees: the root categories with
// their subcategories, published and unpublished, every one or those that
// hold given categories (search), and rebuilds one.
func newCategoryTreesCommand(a *app) *cobra.Command {
	r := categoryTrees()
	// Without --category, Emporix answers every tree, as list does (seen live).
	search := r.searchCommand(a,
		"Reads the category trees, published and unpublished, that hold any of the categories\n"+
			"--category names, each whole; without --category, every tree. --limit counts trees.",
		listField("categoryIds", "category", "a category `ID` whose tree to read (repeatable)"))
	return newResourceCommand(a, r, search, rebuildCommand(a, r))
}

// treeItems returns the categories of a list of category trees (treeCategory).
func treeItems(body json.RawMessage) ([]json.RawMessage, error) {
	roots, err := listItems(body)
	if err != nil {
		return nil, err
	}
	return flattenTrees(roots, 0)
}

// treeRoots returns the root categories of a list of category trees, each a
// treeCategory of level 0, without the categories below them.
func treeRoots(body json.RawMessage) ([]json.RawMessage, error) {
	items, err := treeItems(body)
	return slices.DeleteFunc(items, func(item json.RawMessage) bool {
		var c treeCategory
		return json.Unmarshal(item, &c) != nil || c.Level > 0
	}), err
}

// oneTreeItems returns the categories of one category tree (treeCategory).
func oneTreeItems(body json.RawMessage) ([]json.RawMessage, error) {
	root, err := oneItem(body)
	if err != nil {
		return nil, err
	}
	return flattenTrees(root, 0)
}

// flattenTrees returns each category of nodes followed by its subcategories,
// depth first, as a treeCategory of the given level.
func flattenTrees(nodes []json.RawMessage, level int) ([]json.RawMessage, error) {
	var items []json.RawMessage
	for _, raw := range nodes {
		var node map[string]json.RawMessage
		if err := json.Unmarshal(raw, &node); err != nil || node == nil {
			return nil, errors.New("expected categories as JSON objects")
		}
		var subcategories []json.RawMessage
		if sub, ok := node["subcategories"]; ok && json.Unmarshal(sub, &subcategories) != nil {
			return nil, errors.New("expected subcategories as a JSON array")
		}
		delete(node, "subcategories")
		item, err := json.Marshal(map[string]any{"level": level, "category": node})
		if err != nil {
			return nil, err
		}
		below, err := flattenTrees(subcategories, level+1)
		if err != nil {
			return nil, err
		}
		items = append(append(items, item), below...)
	}
	return items, nil
}
