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

// usageError marks invalid input detected while a command runs.
type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }
func (e usageError) Unwrap() error { return e.err }

// runError marks errors returned by a command's run hooks. Every other error
// comes from Cobra's parsing and validation, which makes it a usage error.
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
	// Error text can contain API-supplied messages, so it is sanitized for the terminal.
	_, _ = fmt.Fprintln(env.Err, "Error:", output.Sanitize(err.Error()))
	code := exitCode(err)
	if code == ExitUsage {
		_, _ = fmt.Fprintln(env.Err, "Run 'epx --help' for usage.")
	}
	return code
}

// forbiddenRefusal is how help names a refusal of the input that Emporix
// answers with 403 Forbidden although the credential is fine, of a write or
// read that declares it so (write.answers.ForbiddenIsRefusal, bodyRead.forbiddenIsRefusal),
// and that its help must name (checkRefusalHelp).
const forbiddenRefusal = "403 Forbidden (exit 1)"

// checkRefusalHelp panics unless the help of cmd, a command that takes a 403
// Forbidden as a refusal of its input, names that refusal (forbiddenRefusal):
// its declaration is then a mistake, which every test of the tree finds.
func checkRefusalHelp(cmd *cobra.Command) {
	if !strings.Contains(cmd.Long, forbiddenRefusal) {
		panic(cmd.Use + ": it takes a 403 as a refusal, which its help does not name")
	}
}

// exitCode maps an error to the exit-code contract. The emporix package
// classifies API failures, so no HTTP status is inspected here.
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

// rootHelp is the long help of the root command.
const rootHelp = "Command-line client for the Emporix Commerce APIs.\n\n" +
	"Credentials are read from the environment:\n" +
	"  EMPORIX_TENANT, EMPORIX_CLIENT_ID, EMPORIX_CLIENT_SECRET (and optionally EMPORIX_API_URL).\n" +
	"Where none of these is set, a profile (`epx profile add`) gives them, its client secret from the\n" +
	"OS keychain: the profile that --profile names, else EPX_PROFILE, else the default profile (`epx\n" +
	"profile use`). The two never mix, and naming a profile while one is set is a usage error.\n" +
	"--tenant overrides the tenant of either. A profile's default site and language apply where a\n" +
	"command needs a site or shows translations and the flag is not given, on the profile's tenant.\n\n" +
	"Exit codes: 0 success, 1 error, 2 invalid usage or configuration,\n" +
	"3 authentication or missing scopes, 4 not found, 5 conflict: a write that Emporix answered with\n" +
	"409 Conflict, as when the item changed since it was read or exists already, or a `units update`\n" +
	"or `catalogs patch` with metadata.version answered with 404 Not Found, as the Unit and Catalog\n" +
	"services answer a stale version. A 403 Forbidden that the help of a command calls a refusal\n" +
	"of its input exits 1."

// newRootCommand builds the command tree writing to out and errOut.
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
	// epx reads no files but a request body (--body of a search, --file of a write), so completion
	// suggests file names only there.
	root.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveNoFileComp)
	// Opt-in: IDs (each read's own) and --site complete from the API.
	if completion, _, err := root.Find([]string{"completion"}); err == nil {
		explainCompletion(completion)
	}
	applyExitPolicy(root)
	return root
}

// topLevelCommands returns the commands directly under epx, in help order.
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

// applyExitPolicy enforces the exit-code contract on the whole tree: groups
// reject unknown subcommands (they print help otherwise), leaves without an
// explicit argument rule accept none, and errors from every run hook are
// marked as runtime errors.
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

// newHelpCommand replaces Cobra's help command, which exits 0 on unknown topics.
// Cobra attaches it at execution time, so applyExitPolicy never wraps it: its
// errors stay unmarked and therefore count as usage errors, which is intended.
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

// errWriter records the first write error, so output that fails to reach the
// terminal (including Cobra's help text) never ends in a success exit code.
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
