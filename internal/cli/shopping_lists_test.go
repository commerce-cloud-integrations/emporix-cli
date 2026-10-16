package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestShoppingListsOfAnUnknownCustomerAreNotFound pins how Emporix answers a
// customer without shopping lists: an empty body, which is not found (exit 4).
func TestShoppingListsOfAnUnknownCustomerAreNotFound(t *testing.T) {
	fake := newFakeEmporix(t)
	fake.route("/shoppinglist/acme1/shopping-lists/c9", respond(http.StatusOK, ``))
	code, stdout, stderr := runCLI(t, fake.env(), "shopping-lists", "get", "c9")
	if code != ExitNotFound || stdout != "" || !strings.Contains(stderr, "empty response: not found") {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// TestShoppingListsSkipMembersThatAreNoList checks that the members of a
// customer that are not lists (no object with items or a name: text, a
// number, null, mixins) are left out of the table.
func TestShoppingListsSkipMembersThatAreNoList(t *testing.T) {
	assertShoppingLists(t, `[{"id":"c1","version":3,"note":null,"label":"x","mixins":{"a":{}},"default":{"items":[]}}]`,
		"CUSTOMER  LISTS\nc1        default (0 items)\n")
}

// TestShoppingListsNameTheCustomerByID checks that id names the customer
// rather than customerId where both are sent, unless it is empty or null, or
// names a list.
func TestShoppingListsNameTheCustomerByID(t *testing.T) {
	assertShoppingLists(t, `[{"id":"c1","customerId":"c0","default":{"items":[]}},`+
		`{"id":null,"customerId":"c2","default":{"items":[]}},{"id":"","customerId":"c3","default":{"items":[]}},`+
		`{"customerId":"c4","id":{"items":[{}]}}]`,
		"CUSTOMER  LISTS\nc1        default (0 items)\nc2        default (0 items)\n"+
			"c3        default (0 items)\nc4        id (1 item)\n")
}

// assertShoppingLists checks that `shopping-lists list` shows body as want.
func assertShoppingLists(t *testing.T, body, want string) {
	t.Helper()
	fake := newFakeEmporix(t)
	fake.route("/shoppinglist/acme1/shopping-lists", respond(http.StatusOK, body))
	code, stdout, stderr := runCLI(t, fake.env(), "shopping-lists", "list")
	if code != ExitOK || stdout != want || stderr != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// TestShoppingListsRejectAMalformedCustomer checks that a list whose items are
// no array, or a customer id that is no text, fails the table (exit 1) and
// names the member.
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
