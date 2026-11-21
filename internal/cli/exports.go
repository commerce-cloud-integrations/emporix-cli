package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

type exportPackage struct {
	Data       string `json:"data"`
	ExportedAt string `json:"exportedAt"`
}

const exportUse = "export"

type export struct {
	short     string
	long      string
	scopes    []string
	selection bodyValue
}

func exportCommand(a *app, path string, e *export) *cobra.Command {
	return reader[exportPackage]{view: view[exportPackage]{
		singular: "export package",
		header:   []string{"DATA", "EXPORTED AT"},
		long:     []string{"DATA"},
		legend: "DATA is the package, base64-encoded and cut in the table; `-o json` prints the package " +
			"whole, as an import on another tenant takes it.",
		row: func(p exportPackage, _ language) []string { return []string{p.Data, p.ExportedAt} },
	}, scopes: e.scopes}.command(a, &command{
		use:   exportUse,
		short: e.short,
		long:  e.long + " Emporix refuses an empty selection (--body with [], exit 1).",
		req:   request{path: path + "/" + exportUse},
		fetch: bodyRead{method: http.MethodPost, body: valueBody(e.selection.require()), fetch: single{items: oneItem}},
	})
}
