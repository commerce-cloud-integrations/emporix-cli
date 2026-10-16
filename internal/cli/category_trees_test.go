package cli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestFlattenTreesRejectsMalformedTrees(t *testing.T) {
	for body, want := range map[string]string{
		`[1]`:                             "expected categories as JSON objects",
		`[null]`:                          "expected categories as JSON objects",
		`[{"id":"r","subcategories":{}}]`: "expected subcategories as a JSON array",
		`{"id":"r"}`:                      "expected a JSON array",
	} {
		if _, err := treeItems(json.RawMessage(body)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("treeItems(%s) error = %v, want %q", body, err, want)
		}
	}
	if _, err := oneTreeItems(json.RawMessage(`[{"id":"r"}]`)); err == nil {
		t.Error("oneTreeItems accepted an array")
	}
}

// treesBody holds two trees; the first has a subcategory with one of its own.
const treesBody = `[{"id":"r1","name":"Root one","subcategories":[{"id":"c1","name":"Child",
	"subcategories":[{"id":"g1","name":"Grandchild"}]}]},{"id":"r2","name":"Root two"}]`

func TestCategoryTreesLimitCountsTrees(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/category/acme1/category-trees", respond(http.StatusOK, treesBody))
	// --limit 1 keeps the first tree whole, its subcategories included.
	code, stdout, stderr := runCLI(t, fake.env(), "category-trees", "list", "--limit", "1")
	want := "ID  CODE  PUBLISHED  NAME\nr1                   Root one\nc1                     Child\n" +
		"g1                       Grandchild\n"
	if code != ExitOK || stdout != want {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	// -o json prints the trees as sent, nesting included.
	code, stdout, _ = runCLI(t, fake.env(), "category-trees", "list", "-o", "json")
	var trees []struct {
		Subcategories []struct {
			Subcategories []struct {
				ID string `json:"id"`
			} `json:"subcategories"`
		} `json:"subcategories"`
	}
	if code != ExitOK || json.Unmarshal([]byte(stdout), &trees) != nil || len(trees) != 2 ||
		len(trees[0].Subcategories) != 1 || len(trees[0].Subcategories[0].Subcategories) != 1 ||
		trees[0].Subcategories[0].Subcategories[0].ID != "g1" {
		t.Fatalf("exit %d, stdout %s", code, stdout)
	}
}

func TestCategoryTreesRejectAMalformedSubcategory(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/category/acme1/category-trees", respond(http.StatusOK,
		`[{"id":"r1","subcategories":[{"id":"c1","subcategories":"none"}]}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "category-trees", "list")
	if code != ExitError || stdout != "" ||
		!strings.Contains(stderr, "decoding category tree response: expected subcategories as a JSON array") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}
