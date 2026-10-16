package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// authStatus is the JSON shape of `auth status`; scopes is never null.
type authStatus struct {
	Tenant           string   `json:"tenant"`
	Scopes           []string `json:"scopes"`
	ExpiresInSeconds int      `json:"expiresInSeconds"`
}

func newAuthCommand(a *app) *cobra.Command {
	auth := &cobra.Command{Use: "auth", Short: "Check Emporix credentials"}
	auth.AddCommand(newAuthStatusCommand(a))
	return auth
}

func newAuthStatusCommand(a *app) *cobra.Command {
	var scopes []string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Request a token and show the tenant and scopes it grants",
		Long: "Requests a service token for the tenant plus the given scopes and shows what was granted.\n" +
			"Emporix grants only requested scopes, so pass --scope to check specific permissions;\n" +
			"a missing scope fails with exit code 3. The token itself is never printed.",
		Example: "  epx auth status --scope product.product_read_unpublished --scope indexing.search_view",
	}
	cmd.Flags().StringSliceVar(&scopes, "scope", nil, "scope the credential must grant (repeatable)")
	printed := addTableOutput(cmd, []string{"TENANT", "EXPIRES IN", "SCOPES"})
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := a.client(cmd.Context(), scopes...)
		if err != nil {
			return err
		}
		info, err := client.Authenticate(cmd.Context())
		if err != nil {
			return err
		}
		status := authStatus{
			Tenant:           info.Tenant,
			Scopes:           append([]string{}, info.Scopes...),
			ExpiresInSeconds: int(info.ExpiresIn.Seconds()),
		}
		return printed.print(cmd.OutOrStdout(), status, func() ([][]string, error) {
			return [][]string{{info.Tenant, info.ExpiresIn.String(), strings.Join(info.Scopes, " ")}}, nil
		})
	}
	return cmd
}
