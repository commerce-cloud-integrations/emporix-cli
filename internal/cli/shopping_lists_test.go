package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestShoppingListsOfAnUnknownCustomerAreNotFound(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/shoppinglist/acme1/shopping-lists/c9", respond(http.StatusOK, ``))
	code, stdout, stderr := runCLI(t, fake.env(), "shopping-lists", "get", "c9")
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "empty response: not found") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestShoppingListsSkipMembersThatAreNoList(t *testing.T) {
	assertShoppingLists(t, `[{"id":"c1","version":3,"note":null,"label":"x","mixins":{"a":{}},"default":{"items":[]}}]`,
		"CUSTOMER  LISTS\nc1        default (0 items)\n")
}

func TestShoppingListsNameTheCustomerByID(t *testing.T) {
	assertShoppingLists(t, `[{"id":"c1","customerId":"c0","default":{"items":[]}},`+
		`{"id":null,"customerId":"c2","default":{"items":[]}},{"id":"","customerId":"c3","default":{"items":[]}},`+
		`{"customerId":"c4","id":{"items":[{}]}}]`,
		"CUSTOMER  LISTS\nc1        default (0 items)\nc2        default (0 items)\n"+
			"c3        default (0 items)\nc4        id (1 item)\n")
}

func assertShoppingLists(t *testing.T, body, want string) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route("/shoppinglist/acme1/shopping-lists", respond(http.StatusOK, body))
	code, stdout, stderr := runCLI(t, fake.env(), "shopping-lists", "list")
	if code != ExitOK || stdout != want || stderr != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestShoppingListsRejectAMalformedCustomer(t *testing.T) {
	for member, body := range map[string]string{
		`member "default"`:    `[{"id":"c1","default":{"name":"default","items":"x"}}]`,
		`member "id"`:         `[{"id":7,"default":{"items":[]}}]`,
		`member "customerId"`: `[{"customerId":{},"default":{"items":[]}}]`,
	} {
		fake := newFakeEmporix(t)
		fake.route("/shoppinglist/acme1/shopping-lists", respond(http.StatusOK, body))
		code, stdout, stderr := runCLI(t, fake.env(), "shopping-lists", "list")
		if code != ExitError || stdout != "" || !strings.Contains(stderr, member) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", body, code, stdout, stderr)
		}
	}
}
