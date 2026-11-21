package cli

import (
	"net/http"
	"strings"
	"testing"
)

const (
	shippingZoneJSON   = `{"id":"nl","name":{"en":"Netherlands"},"shipTo":[{"country":"NL"}],"default":true}`
	shippingMethodJSON = `{"id":"post","name":"Post","active":true,"fees":[{"minOrderValue":{"amount":0,` +
		`"currency":"EUR"},"cost":{"amount":5,"currency":"EUR"}}],"shippingTaxCode":"STANDARD"}`
)

func shippingWrites() []writeCase {
	manage := []string{shippingManageScope}
	site := []string{"--site", "AcmeStore_NL"}
	return []writeCase{
		{
			args: append([]string{"shipping", "zones", "create"}, site...), file: shippingZoneJSON,
			method: http.MethodPost, path: "/shipping/acme1/AcmeStore_NL/zones", scopes: manage,
			status: http.StatusCreated, response: `{"id":"nl"}`, want: "{\n  \"id\": \"nl\"\n}\n",
		},
		{
			args: []string{"shipping", "zones", "update", "nl 1", "--site", "AcmeStore_NL", "--content-language",
				"en-NL"},
			file:   `{"id":"nl 1","name":"Netherlands","shipTo":[{"country":"NL","postalCode":"10*"}]}`,
			method: http.MethodPut, path: "/shipping/acme1/AcmeStore_NL/zones/nl%201", scopes: manage, lang: "en-NL",
			status: http.StatusOK,
		},
		{
			args: append([]string{"shipping", "zones", "patch", "nl", "--content-language", "nl"}, site...),
			file: `{"name":"Nederland"}`, method: http.MethodPatch, path: "/shipping/acme1/AcmeStore_NL/zones/nl",
			scopes: manage, lang: "nl", status: http.StatusOK,
		},
		{
			args:   append([]string{"shipping", "zones", "delete", "nl"}, site...),
			method: http.MethodDelete, path: "/shipping/acme1/AcmeStore_NL/zones/nl", scopes: manage,
			confirmBy: "nl", status: http.StatusNoContent,
		},
		{
			args: append([]string{"shipping", "zones", "methods", "create", "nl 1"}, site...), file: shippingMethodJSON,
			method: http.MethodPost, path: "/shipping/acme1/AcmeStore_NL/zones/nl%201/methods", scopes: manage,
			status: http.StatusCreated, response: `{"id":"post"}`, want: "{\n  \"id\": \"post\"\n}\n",
		},
		{
			args: append([]string{"shipping", "zones", "methods", "update", "nl", "post 1"}, site...),
			file: shippingMethodJSON, method: http.MethodPut,
			path: "/shipping/acme1/AcmeStore_NL/zones/nl/methods/post%201", scopes: manage,
			status: http.StatusOK,
		},
		{
			args: append([]string{"shipping", "zones", "methods", "patch", "nl", "post"}, site...),
			file: `{"active":false}`, method: http.MethodPatch,
			path: "/shipping/acme1/AcmeStore_NL/zones/nl/methods/post", scopes: manage, status: http.StatusOK,
		},
		{
			args:   append([]string{"shipping", "zones", "methods", "delete", "nl", "post"}, site...),
			method: http.MethodDelete, path: "/shipping/acme1/AcmeStore_NL/zones/nl/methods/post", scopes: manage,
			confirmBy: "post", status: http.StatusNoContent,
		},
		{
			args: append([]string{"shipping", "groups", "create"}, site...),
			file: `{"id":"vip","description":{"en":"Free shipping"}}`, method: http.MethodPost,
			path: "/shipping/acme1/AcmeStore_NL/groups", scopes: manage, status: http.StatusCreated,
			response: `{"id":"vip"}`, want: "{\n  \"id\": \"vip\"\n}\n",
		},
		{
			args: append([]string{"shipping", "groups", "update", "vip"}, site...),
			file: `{"id":"vip","description":{"en":"Free shipping over 50 EUR"}}`, method: http.MethodPut,
			path: "/shipping/acme1/AcmeStore_NL/groups/vip", scopes: manage, status: http.StatusOK,
			conflict: unknownGroup,
		},
		{
			args:   append([]string{"shipping", "groups", "delete", "vip"}, site...),
			method: http.MethodDelete, path: "/shipping/acme1/AcmeStore_NL/groups/vip", scopes: manage,
			confirmBy: "vip", status: http.StatusNoContent,
		},
		{
			args: append([]string{"shipping", "customers", "create"}, site...),
			file: `{"customerId":"c1","groupId":"vip"}`, method: http.MethodPost,
			path: "/shipping/acme1/AcmeStore_NL/cgrelations", scopes: manage, status: http.StatusCreated,
			response: `{"id":"c1"}`, want: "{\n  \"id\": \"c1\"\n}\n",
		},
		{
			args: append([]string{"shipping", "customers", "update", "c1"}, site...),
			file: `{"customerId":"c1","groupId":"standard"}`, method: http.MethodPut,
			path: "/shipping/acme1/AcmeStore_NL/cgrelations/c1", scopes: manage, status: http.StatusOK,
			conflict: noGroup,
		},
		{
			args:   append([]string{"shipping", "customers", "delete", "c1"}, site...),
			method: http.MethodDelete, path: "/shipping/acme1/AcmeStore_NL/cgrelations/c1", scopes: manage,
			confirmBy: "c1", status: http.StatusNoContent,
		},
	}
}

func TestChildWritesNameTheirArguments(t *testing.T) {
	const (
		zone   = "SHIPPING_ZONE_ID is the ID that `epx shipping zones list --site S` shows"
		method = " ID is the ID that `epx shipping zones methods list SHIPPING_ZONE_ID --site S` shows; an unknown " +
			"one is not found.\n\n"
		time = "DELIVERY_TIME_ID is the ID that `epx delivery-times list` shows"
		slot = " ID is the ID that `epx delivery-times slots list DELIVERY_TIME_ID` shows; an unknown one is not " +
			"found.\n\n"
		unknown = "; an unknown one is not found."
	)
	root := newTestRoot()
	for _, tt := range []struct {
		args []string
		use  string
		want string
	}{
		{args: []string{"shipping", "zones", "methods", "create"}, use: "create SHIPPING_ZONE_ID", want: zone + unknown},
		{args: []string{"shipping", "zones", "methods", "update"}, use: "update SHIPPING_ZONE_ID ID",
			want: zone + unknown + method},
		{args: []string{"shipping", "zones", "methods", "patch"}, use: "patch SHIPPING_ZONE_ID ID",
			want: zone + unknown + method},
		{args: []string{"shipping", "zones", "methods", "delete"}, use: "delete SHIPPING_ZONE_ID ID",
			want: zone + unknown + method},
		{args: []string{"delivery-times", "slots", "create"}, use: "create DELIVERY_TIME_ID", want: time + unknown + "\n\n"},
		{args: []string{"delivery-times", "slots", "update"}, use: "update DELIVERY_TIME_ID ID",
			want: time + unknown + slot},
		{args: []string{"delivery-times", "slots", "patch"}, use: "patch DELIVERY_TIME_ID ID",
			want: time + unknown + slot},
		{args: []string{"delivery-times", "slots", "delete"}, use: "delete DELIVERY_TIME_ID ID",
			want: time + unknown + " ID is the ID that `epx delivery-times slots list DELIVERY_TIME_ID` shows; an " +
				"unknown one is answered as one that Emporix deleted (204 No Content).\n\n"},
		{args: []string{"delivery-times", "slots", "delete-all"}, use: "delete-all DELIVERY_TIME_ID",
			want: time + unknown + "\n\n"},
	} {
		cmd, _, err := root.Find(tt.args)
		if err != nil || cmd.Use != tt.use || !strings.Contains(cmd.Long, tt.want) {
			t.Errorf("%q (%v): use %q, help %q; want %q and %q", tt.args, err, cmd.Use, cmd.Long, tt.use, tt.want)
		}
	}
}

func TestShippingCustomerWritesHelp(t *testing.T) {
	assertHelpHolds(t, map[string][]string{
		"shipping customers create": {"unknown ids are assigned too"},
		"shipping customers update": {"customerId in the JSON is ignored", noGroup},
		"shipping customers delete": {"without a shipping group, or an unknown one, as one that"},
		"shipping groups delete":    {"customers assigned to it stay assigned"},
	})
}
