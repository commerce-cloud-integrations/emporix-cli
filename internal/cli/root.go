// Package cli builds the epx command tree and maps failures to exit codes.
package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/config"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

// Exit codes returned by Run.
const (
	ExitOK       = 0
	ExitError    = 1
	ExitUsage    = 2
	ExitAuth     = 3 // credential rejected, other tenant or missing scopes
	ExitNotFound = 4
	ExitConflict = 5 // a write that Emporix answered with 409 Conflict
)

// Env is the process context the CLI runs in.
type Env struct {
	In       io.Reader // a body given as - and a write's confirmation (os.Stdin when nil)
	Out      io.Writer
	Err      io.Writer
	Getenv   func(string) string
	Terminal bool // In is a terminal: a write asks for confirmation there, and needs --yes otherwise
	// OutTerminal says that Out is a terminal, to which a read never writes
	// the file that Emporix answers with, such as a PDF (fileRead).
	OutTerminal bool
	Profiles    *ProfileEnv // where profiles are kept; only a run whose command reads no profile may do without
}

// ProfileEnv is what profiles need of the system.
type ProfileEnv struct {
	ConfigDir func() (string, error) // the user's configuration directory (os.UserConfigDir)
	Keychain  config.Keychain        // holds the client secrets of profiles
	// ReadHidden reads a line from the terminal In without showing it: a
	// client secret (ReadHidden in this package).
	ReadHidden func(context.Context) (string, error)
}

type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }
func (e usageError) Unwrap() error { return e.err }

type runError struct{ err error }

func (e runError) Error() string { return e.err.Error() }
func (e runError) Unwrap() error { return e.err }

// Run executes the CLI with args (excluding the program name) and returns the exit code.
func Run(ctx context.Context, args []string, env Env) int {
	a := &app{
		getenv: env.Getenv, terminal: env.Terminal, outTerminal: env.OutTerminal, version: "unknown",
		goVersion: "unknown", completionTimeout: defaultCompletionTimeout,
	}
	if env.Profiles != nil {
		a.ProfileEnv = *env.Profiles
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		a.version, a.goVersion = cmp.Or(info.Main.Version, a.version), info.GoVersion
	}
	out := &errWriter{w: env.Out}
	//nolint:contextcheck // RunE closures use cmd.Context(), which Cobra derives from ExecuteContext(ctx).
	root := newRootCommand(a, out, env.Err)
	root.SetIn(env.In)
	root.SetArgs(args)
	err := root.ExecuteContext(ctx)
	if err == nil && out.err != nil {
		err = runError{fmt.Errorf("writing output: %w", out.err)}
	}
	if err == nil {
		return ExitOK
	}
	_, _ = fmt.Fprintln(env.Err, "Error:", output.Sanitize(err.Error()))
	code := exitCode(err)
	if code == ExitUsage {
		_, _ = fmt.Fprintln(env.Err, "Run 'epx --help' for usage.")
	}
	return code
}

const forbiddenRefusal = "403 Forbidden (exit 1)"

func checkRefusalHelp(cmd *cobra.Command) {
	if !strings.Contains(cmd.Long, forbiddenRefusal) {
		panic(cmd.Use + ": it takes a 403 as a refusal, which its help does not name")
	}
}

func exitCode(err error) int {
	switch {
	case !errors.As(err, new(runError)), errors.As(err, new(usageError)), errors.Is(err, emporix.ErrInvalidInput):
		return ExitUsage
	case errors.Is(err, emporix.ErrAuth):
		return ExitAuth
	case errors.Is(err, emporix.ErrNotFound):
		return ExitNotFound
	case errors.Is(err, emporix.ErrConflict):
		return ExitConflict
	default:
		return ExitError
	}
}

const rootHelp = "Command-line client for the Emporix Commerce APIs.\n\n" +
	"Credentials are read from the environment:\n" +
	"  EMPORIX_TENANT, EMPORIX_CLIENT_ID, EMPORIX_CLIENT_SECRET (and optionally EMPORIX_API_URL).\n" +
	"Where none of these is set, a profile (`epx profile add`) gives them, its client secret from the " +
	"OS keychain: the profile that --profile names, else EPX_PROFILE, else the default profile (`epx " +
	"profile use`). The two never mix, and naming a profile while one is set is a usage error. " +
	"--tenant overrides the tenant of either. A profile's default site and language apply where a " +
	"command needs a site or shows translations and the flag is not given, on the profile's tenant.\n\n" +
	"Exit codes: 0 success, 1 error, 2 invalid usage or configuration, " +
	"3 authentication or missing scopes, 4 not found, 5 conflict: a write that Emporix answered with " +
	"409 Conflict, as when the item changed since it was read or exists already, or a `units update` " +
	"or `catalogs patch` with metadata.version answered with 404 Not Found, as the Unit and Catalog " +
	"services answer a stale version. A 403 Forbidden that the help of a command calls a refusal " +
	"of its input exits 1."

func newRootCommand(a *app, out, errOut io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "epx",
		Short:         "Command-line client for the Emporix Commerce APIs",
		Long:          rootHelp,
		SilenceErrors: true,
		SilenceUsage:  true,
		// The one persistent hook of the tree: Cobra runs only the nearest, so no command below may
		// define another (TestOnlyTheRootHasAPersistentHook).
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			completeFlagsFor(a, cmd)
			return a.applyProfileDefaults(cmd)
		},
	}
	flags := root.PersistentFlags()
	flags.Var(&a.tenant, "tenant", "Emporix tenant (default from "+envTenant+", else the profile)")
	flags.Var(&a.profile, profileFlag, "profile to use (default from "+envProfile+", else the default profile)")
	flags.DurationVar(&a.timeout, "timeout", defaultTimeout, "timeout for each HTTP request")
	root.SetHelpCommand(newHelpCommand())
	root.AddCommand(topLevelCommands(a)...)
	// Cobra would add `completion` only while executing; adding it now puts it
	// under the exit policy (an unknown shell is a usage error). It captures the
	// output writer when created, so the writers are set first.
	root.SetOut(out)
	root.SetErr(errOut)
	root.InitDefaultCompletionCmd()
	root.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveNoFileComp)
	if completion, _, err := root.Find([]string{"completion"}); err == nil {
		explainCompletion(completion)
	}
	wrapHelp(root)
	applyExitPolicy(root)
	return root
}

func topLevelCommands(a *app) []*cobra.Command {
	return []*cobra.Command{
		newVersionCommand(a),
		newAuthCommand(a),
		newProfileCommand(a),
		newAPICommand(a),
		newSitesCommand(a),
		newProductsCommand(a),
		newCategoriesCommand(a),
		newCategoryTreesCommand(a),
		newCatalogsCommand(a),
		newUnitsCommand(a),
		newBrandsCommand(a),
		newLabelsCommand(a),
		newMediaCommand(a),
		newSchemasCommand(a),
		newCustomEntitiesCommand(a),
		newProductTemplatesCommand(a),
		newPricesCommand(a),
		newPriceModelsCommand(a),
		newPriceListsCommand(a),
		newFeesCommand(a),
		newShippingCommand(a),
		newDeliveryTimesCommand(a),
		newPaymentModesCommand(a),
		newPaymentTransactionsCommand(a),
		newCustomersCommand(a),
		newOrdersCommand(a),
		newInvoiceJobsCommand(a),
		newReturnsCommand(a),
		newQuotesCommand(a),
		newShoppingListsCommand(a),
		newCartsCommand(a),
		newCouponsCommand(a),
		newRewardPointsCommand(a),
		newIAMCommand(a),
		newSessionContextsCommand(a),
		newWebhooksCommand(a),
		newLegalEntitiesCommand(a),
		newCustomerLocationsCommand(a),
		newContactAssignmentsCommand(a),
		newSegmentsCommand(a),
		newVendorsCommand(a),
		newVendorLocationsCommand(a),
		newAvailabilityCommand(a),
		newIndexingCommand(a),
		newSearchCommand(a),
		newAICommand(a),
		newConfigurationsCommand(a),
		newSequentialIDsCommand(a),
		newChangelogCommand(a),
		newImportsCommand(a),
		newCurrenciesCommand(a),
		newExchangeRatesCommand(a),
		newCountriesCommand(a),
		newRegionsCommand(a),
		newTaxesCommand(a),
	}
}

func applyExitPolicy(cmd *cobra.Command) {
	if cmd.HasSubCommands() {
		cmd.Args = cobra.NoArgs
		cmd.RunE = func(c *cobra.Command, _ []string) error { return c.Help() }
	} else if cmd.Args == nil {
		cmd.Args = cobra.NoArgs
	}
	for _, hook := range []*func(*cobra.Command, []string) error{
		&cmd.PersistentPreRunE, &cmd.PreRunE, &cmd.RunE, &cmd.PostRunE, &cmd.PersistentPostRunE,
	} {
		if run := *hook; run != nil {
			*hook = func(c *cobra.Command, args []string) error {
				if err := run(c, args); err != nil {
					return runError{err}
				}
				return nil
			}
		}
	}
	for _, sub := range cmd.Commands() {
		applyExitPolicy(sub)
	}
}

// Cobra's own help command exits 0 on unknown topics. Cobra attaches the help
// command at execution time, so applyExitPolicy never wraps it: its errors
// stay unmarked and therefore count as usage errors.
func newHelpCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "help [command]",
		Short:             "Help about any command",
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: completeCommandNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, rest, err := cmd.Root().Find(args)
			if err != nil || len(rest) > 0 {
				return fmt.Errorf("unknown help topic %q", strings.Join(args, " "))
			}
			return target.Help()
		},
	}
}

type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) Write(p []byte) (int, error) {
	n, err := e.w.Write(p)
	if err != nil && e.err == nil {
		e.err = err
	}
	return n, err
}
