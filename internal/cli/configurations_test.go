package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestConfigurationsListShortensValues checks that tables show settings' values
// as sent, compacted and cut, and that -o json prints them whole.
func TestConfigurationsListShortensValues(t *testing.T) {
	fake := newFakeEmporix(t)
	// The service ignores paging and answers every page with all settings.
	fake.route("/configuration/acme1/configurations", respond(http.StatusOK, `[
		{"key":"project_lang","value":"[{\"id\":\"nl\",\"default\":true}]","version":2,"secured":false,"readOnly":false},
		{"key":"taxConfiguration","value":{"taxClasses":{"FULL":21, "ZERO":0}},"version":1,"secured":false},
		{"key":"transferKey","value":"encoding-sentinel","version":1,"secured":false},
		{"key":"socialLogin","value":{"provider":"social-sentinel"},"version":0,"secured":true},
		{"key":"integration","value":{"clientSecret":"nested-sentinel","url":"u"},"version":1},
		{"key":"scorecardskey","value":null,"version":1},
		{"key":"storefront.long","value":"`+strings.Repeat("a", 70)+`","version":1}]`))
	code, stdout, stderr := runCLI(t, fake.env(), "configurations", "list")
	// Secured and credential-named settings show their values as sent; a null value stays
	// empty (as null in JSON), and a missing secured flag stays empty.
	want := strings.Join([]string{
		"KEY               VALUE                                                         VERSION  SECURED  READ ONLY",
		`project_lang      [{"id":"nl","default":true}]                                  2        false    false`,
		`taxConfiguration  {"taxClasses":{"FULL":21,"ZERO":0}}                           1        false    `,
		"transferKey       encoding-sentinel                                             1        false    ",
		`socialLogin       {"provider":"social-sentinel"}                                0        true     `,
		`integration       {"clientSecret":"nested-sentinel","url":"u"}                  1                 `,
		"scorecardskey                                                                   1                 ",
		"storefront.long   " + strings.Repeat("a", 59) + "…  1                 ",
	}, "\n") + "\n"
	if code != ExitOK || stdout != want {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	code, stdout, _ = runCLI(t, fake.env(), "configurations", "list", "-o", "json")
	if code != ExitOK || !strings.Contains(stdout, `"provider": "social-sentinel"`) ||
		!strings.Contains(stdout, `"FULL": 21`) || !strings.Contains(stdout, strings.Repeat("a", 70)) {
		t.Fatalf("exit %d; JSON must print every setting whole:\n%s", code, stdout)
	}
}
