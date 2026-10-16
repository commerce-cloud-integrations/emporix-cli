package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestJSONPrintsResponsesAsSent checks that -o json prints what reads that
// once hid values print as Emporix sends them: changed values in a page of
// changes, card tokens and provider requests and responses kept as JSON text,
// webhook headers and destination URLs; and, whole, what tables leave out or
// cut: the webhook dashboard's token and an export package's data.
func TestJSONPrintsResponsesAsSent(t *testing.T) {
	for _, tt := range []struct {
		args []string
		path string
		body http.HandlerFunc
		want []string
	}{
		{
			args: []string{"changelog", "list"},
			path: "/changelog/acme1/changelogs",
			body: respond(http.StatusOK, `{"items":[{"at":"2026-10-01T13:01:29.123Z","type":"update",
				"entity":"site","entityId":"AcmeStore_NL","paths":{"mixins.api.key.value":{"before":"S1","after":"S2"}}}],
				"page":1,"size":100,"totalElements":1,"totalPages":1}`),
			want: []string{`"before": "S1"`, `"after": "S2"`},
		},
		{
			args: []string{"payment-transactions", "list"},
			path: "/payment-gateway/acme1/transactions",
			body: firstPage(transactionsBody),
			want: []string{
				`"creditCardToken": "S1"`, `"remoteToken": "S5"`,
				`"remoteResponse": "{\"transaction\":{\"token\":\"S3\",\"succeeded\":true,\"gateway_token\":\"S4\"}}"`,
			},
		},
		{
			args: []string{"webhooks", "configurations", "get", "erp"},
			path: "/webhook/acme1/config/erp",
			body: respond(http.StatusOK, `{"code":"erp","provider":"HTTP","configuration":{
				"destinationUrl":"https://erp.example.com/hook?code=S6#f","headers":{"Authorization":{"value":"Bearer S1"}}}}`),
			want: []string{`"destinationUrl": "https://erp.example.com/hook?code=S6#f"`, `"value": "Bearer S1"`},
		},
		{
			args: []string{"webhooks", "dashboard-access", "--" + writeScopeFlag},
			path: "/webhook/acme1/dashboard-access",
			body: respond(http.StatusOK, `{"url":"https://app.svix.example/login#key=S1","token":"appsk_S2"}`),
			want: []string{`"url": "https://app.svix.example/login#key=S1"`, `"token": "appsk_S2"`},
		},
		{
			args: []string{"custom-entities", "export", "--type", "CAR", "--" + writeScopeFlag},
			path: "/schema/acme1/custom-entities/export",
			body: respond(http.StatusOK, `{"data":"`+exportedData+`"}`),
			want: []string{`"data": "` + exportedData + `"`},
		},
	} {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route(tt.path, tt.body)
			code, stdout, stderr := runCLI(t, fake.env(), append(tt.args, "-o", "json")...)
			for _, want := range tt.want {
				if code != ExitOK || !strings.Contains(stdout, want) {
					t.Fatalf("exit %d, stderr %q, stdout missing %s:\n%s", code, stderr, want, stdout)
				}
			}
		})
	}
}
