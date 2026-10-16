package cli

import (
	"net/http"
	"testing"
)

// TestSitesMixinsListEmptyBody checks that an empty body of the mixins of a site is no mixins, and
// that an unknown site, which Emporix answers with 404 on the mixins path, is not found.
func TestSitesMixinsListEmptyBody(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		args     []string
		wantCode int
		wantOut  string
	}{
		{
			name:     "a site without mixins lists none",
			status:   http.StatusOK,
			args:     []string{"sites", "mixins", "list", "CH"},
			wantCode: ExitOK,
			wantOut:  "NAME  VALUE\n",
		},
		{
			name:     "the same site as -o json is the empty object",
			status:   http.StatusOK,
			args:     []string{"sites", "mixins", "list", "CH", "-o", "json"},
			wantCode: ExitOK,
			wantOut:  "{}\n",
		},
		{
			name:     "an unknown site is not found",
			status:   http.StatusNotFound,
			body:     `{"message":"not found"}`,
			args:     []string{"sites", "mixins", "list", "unknown"},
			wantCode: ExitNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeEmporix(t)
			fake.route("/site/acme1/sites/"+tt.args[3]+"/mixins", respond(tt.status, tt.body))
			code, stdout, stderr := runCLI(t, fake.env(), tt.args...)
			if code != tt.wantCode || stdout != tt.wantOut {
				t.Fatalf("exit %d, stderr %q, stdout %q; want exit %d, stdout %q", code, stderr, stdout, tt.wantCode, tt.wantOut)
			}
		})
	}
}
