package cli

import (
	"errors"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

func newAPICommand(a *app) *cobra.Command {
	api := &cobra.Command{Use: "api", Short: "Send requests to any Emporix API path"}
	api.AddCommand(newAPIGetCommand(a),
		newAPIWriteCommand(a, http.MethodPost, "/brand/brands --scope brand.brand_manage --file brand.json"),
		newAPIWriteCommand(a, http.MethodPut,
			"/product/{tenant}/products/100200 --scope product.product_manage --file product.json --dry-run"),
		newAPIWriteCommand(a, http.MethodPatch,
			"/site/{tenant}/sites/AcmeStore_NL --scope site.site_manage --file site.json"),
		newAPIWriteCommand(a, http.MethodDelete, "/product/{tenant}/products/100200 --scope product.product_manage --yes"))
	return api
}

func newAPIGetCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get PATH",
		Short: "GET an API path and print the JSON response",
		Long: "Sends GET to PATH on the configured API host and prints the response body as JSON\n" +
			"(YAML with -o yaml) with sorted keys and every value as sent.\n" +
			"{tenant} in PATH is replaced with the tenant; Accept-Language: * requests every translation.\n" +
			"Query values given with -p replace inline ones. Request every scope the endpoint needs\n" +
			"with --scope: Emporix grants only requested scopes and may silently omit data otherwise.",
		Example: "  epx api get /site/{tenant}/sites\n" +
			"  epx api get /product/{tenant}/products -p pageSize=5 --scope product.product_read_unpublished",
		Args: cobra.ExactArgs(1),
	}
	scopes, query := addAPIFlags(cmd)
	format := addResponseOutput(cmd)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		client, err := a.client(c.Context(), *scopes...)
		if err != nil {
			return err
		}
		body, err := client.Get(c.Context(), args[0], url.Values(query))
		if err != nil {
			return err
		}
		return output.Write(c.OutOrStdout(), format.chosen(), body)
	}
	return cmd
}

// newAPIWriteCommand returns `api post|put|patch|delete PATH`, a write of
// method to any API path, which may contain {tenant} and a query, with the
// query values of -p, the scopes of --scope and the JSON body of --file, if
// any; its example takes example as the arguments.
func newAPIWriteCommand(a *app, method, example string) *cobra.Command {
	use := strings.ToLower(method)
	cmd := &cobra.Command{
		Use:   use + " PATH",
		Short: "Send " + method + " to an API path and print the JSON response",
		Long: "Sends " + method + " to PATH on the configured API host, with the JSON body that --file gives\n" +
			"(none without it), and prints the response body as JSON (YAML with -o yaml) with sorted\n" +
			"keys and every value as sent (nothing for an empty one). {tenant} in PATH is replaced with\n" +
			"the tenant; query values given with -p replace inline ones. It sends no other header, such\n" +
			"as the Content-Language that the Unit Service requires to create or update a unit.",
		Example: "  epx api " + use + " " + example,
		Args:    cobra.ExactArgs(1),
	}
	scopes, query := addAPIFlags(cmd)
	// A delete is confirmed by typing PATH: what it deletes is unknown, and may be many items.
	s := &sender{a: a, method: method, body: &payload{optional: true},
		confirm: confirmationFor(method, "PATH", deletion{})}
	s.bind(cmd)
	cmd.Long += "\n\nRequests only the scopes given with --scope besides the tenant: name every write scope\n" +
		"that the endpoint needs, as Emporix grants only requested scopes."
	cmd.RunE = func(c *cobra.Command, args []string) error {
		return s.send(c, *scopes, &emporix.Write{Path: args[0], Query: url.Values(query)}, nil, subject{word: args[0]})
	}
	return cmd
}

// addAPIFlags adds to cmd, a command of api, --scope and -p, and returns the
// scopes and the query values that they give.
func addAPIFlags(cmd *cobra.Command) (scopes *[]string, query queryParams) {
	scopes, query = new([]string), queryParams{}
	cmd.Flags().StringSliceVar(scopes, "scope", nil, "OAuth scope to request besides the tenant (repeatable)")
	cmd.Flags().VarP(query, "param", "p", "query parameter as key=value (repeatable)")
	return scopes, query
}

// queryParams are --param values; it implements pflag.Value, so a value that
// is not key=value is rejected while flags are parsed.
type queryParams url.Values

// Set implements pflag.Value, adding one key=value.
func (q queryParams) Set(param string) error {
	key, value, ok := strings.Cut(param, "=")
	if !ok || key == "" {
		return errors.New("use key=value") // the flag parser names the flag and the value
	}
	url.Values(q).Add(key, value)
	return nil
}

func (q queryParams) String() string { return url.Values(q).Encode() }

// GetSlice returns the parameters as key=value pairs; it lets completion offer
// the repeatable --param again (cobra.SliceValue).
func (q queryParams) GetSlice() []string {
	var params []string
	for _, key := range slices.Sorted(maps.Keys(q)) {
		for _, value := range q[key] {
			params = append(params, key+"="+value)
		}
	}
	return params
}

// Type implements pflag.Value.
func (q queryParams) Type() string { return "key=value" }
