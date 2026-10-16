package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// exportPackage is the part of an export package shown in tables: what
// Emporix exported, base64-encoded, and when.
type exportPackage struct {
	Data       string `json:"data"`
	ExportedAt string `json:"exportedAt"`
}

// exportUse names the reads of an export package.
const exportUse = "export"

// export declares the read of an export package (exportCommand).
type export struct {
	short     string    // summary of the command
	long      string    // what it exports; the empty selection, package and scope sentences follow
	scopes    []string  // OAuth scopes besides the tenant
	selection bodyValue // what to export, which its flag selects: the whole request body (required)
}

// exportCommand returns `export`, which reads the export package of what e's
// selection selects: Emporix answers a POST to path/export, whose body is the
// selection, with the package, which an import on another tenant takes as its
// body. Emporix refuses an empty selection, so its flag is required unless
// --body gives the body (exit 2 before anything is sent).
func exportCommand(a *app, path string, e *export) *cobra.Command {
	return reader[exportPackage]{view: view[exportPackage]{
		singular: "export package",
		header:   []string{"DATA", "EXPORTED AT"},
		long:     []string{"DATA"},
		legend: "DATA is the package, base64-encoded and cut in the table; -o json prints the package\n" +
			"whole, as an import on another tenant takes it.",
		row: func(p exportPackage, _ language) []string { return []string{p.Data, p.ExportedAt} },
	}, scopes: e.scopes}.command(a, &command{
		use:   exportUse,
		short: e.short,
		long:  e.long + "\nEmporix refuses an empty selection (--body with [], exit 1).",
		req:   request{path: path + "/" + exportUse},
		fetch: bodyRead{method: http.MethodPost, body: valueBody(e.selection.require()), fetch: single{items: oneItem}},
	})
}
