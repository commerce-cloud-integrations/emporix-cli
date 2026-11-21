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

type treeCategory struct {
	Level    int      `json:"level"`
	Category category `json:"category"`
}

func categoryTrees() resource[treeCategory] {
	return resource[treeCategory]{
		use:   "category-trees",
		short: "Read and rebuild category trees",
		long: "Reads category trees, published and unpublished: each root category followed by its " +
			"subcategories, depth first, in the order Emporix sends them; NAME is indented by the " +
			"category's level. get takes the id of a root category: Emporix answers another " +
			"category's id as not found. list's --limit counts trees, each shown whole. list's --type " +
			"lists only the trees of one category type: STANDARD (regular categories) or " +
			"CLASSIFICATION (categories that define classification mixins, the attributes that the " +
			"products in them take). A move of a category shows in its tree a few seconds later; a " +
			"root category moved below another may keep its old tree until it is deleted.",
		path:  "/category/{tenant}/category-trees",
		query: withUnpublished(),
		listParams: []paramFlag{{
			param: typeParam, flag: typeParam, choices: []string{"STANDARD", "CLASSIFICATION"},
			usage: "list only the trees of this category `TYPE`",
		}},
		scopes: []string{"category.category_read_unpublished"},
		read:   (*emporix.Client).GetUnpaged,
		items:  treeItems,
		item:   oneTreeItems,
		roots:  treeRoots,
		view: view[treeCategory]{
			singular:  "category tree",
			header:    []string{columnID, columnCode, columnPublished, columnName},
			localized: true,
			row: func(c treeCategory, lang language) []string {
				name := strings.TrimLeftFunc(output.Sanitize(c.Category.displayName(lang)), unicode.IsSpace)
				return []string{
					c.Category.ID, c.Category.Code, formatBool(c.Category.Published), strings.Repeat("  ", c.Level) + name,
				}
			},
		},
	}
}

func newCategoryTreesCommand(a *app) *cobra.Command {
	r := categoryTrees()
	search := r.searchCommand(a,
		"Reads the category trees, published and unpublished, that hold any of the categories "+
			"--category names, each whole; without --category, every tree. --limit counts trees.",
		listField("categoryIds", "category", "a category `ID` whose tree to read (repeatable)"))
	return newResourceCommand(a, r, search, rebuildCommand(a, r))
}

func treeItems(body json.RawMessage) ([]json.RawMessage, error) {
	roots, err := listItems(body)
	if err != nil {
		return nil, err
	}
	return flattenTrees(roots, 0)
}

func treeRoots(body json.RawMessage) ([]json.RawMessage, error) {
	items, err := treeItems(body)
	return slices.DeleteFunc(items, func(item json.RawMessage) bool {
		var c treeCategory
		return json.Unmarshal(item, &c) != nil || c.Level > 0
	}), err
}

func oneTreeItems(body json.RawMessage) ([]json.RawMessage, error) {
	root, err := oneItem(body)
	if err != nil {
		return nil, err
	}
	return flattenTrees(root, 0)
}

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
