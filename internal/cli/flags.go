package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The flags a read adds to its command (list filters: listFilters.register;
// output: addTableOutput).

// requireFlag makes --flag of cmd, which cmd has just added, required:
// without it, the command fails as its flags are parsed (exit 2), and nothing
// is sent. Its usage says so (noteRequired).
func requireFlag(cmd *cobra.Command, flag string) {
	// MarkFlagRequired fails only for an undefined flag, and the flag was just added.
	_ = cmd.MarkFlagRequired(flag)
	noteRequired(cmd, flag, "required")
}

// requireFlagUnless makes --flag of cmd, which cmd has just added, required
// unless --other is given, as requireFlag makes a flag required.
func requireFlagUnless(cmd *cobra.Command, flag, other string) {
	cmd.MarkFlagsOneRequired(other, flag)
	noteRequired(cmd, flag, "required without --"+other)
}

// noteRequired adds note to the usage of --flag of cmd, in the closing
// parenthesis that ends it, if any.
func noteRequired(cmd *cobra.Command, flag, note string) {
	given := cmd.Flags().Lookup(flag)
	if usage, closed := strings.CutSuffix(given.Usage, ")"); closed {
		given.Usage = usage + "; " + note + ")"
	} else {
		given.Usage += " (" + note + ")"
	}
}

// languageFlag is the flag that names a language: of the localized text in a
// table, or of a file that Emporix answers with (fileRead).
const languageFlag = "language"

// addLanguageFlag adds --language to a read whose table shows translations;
// it takes the default language of the profile in effect.
func addLanguageFlag(cmd *cobra.Command, lang *language) {
	cmd.Flags().Var(lang, languageFlag,
		"language of localized table text; a missing translation shows another, labeled (-o json prints all)")
	takeProfileDefault(cmd, languageFlag, languageFlag)
}

// siteFlag is the flag that names a site: the site of a read whose request
// names it (addSiteFlag), the sites of a search's body (siteCodeField,
// siteCodesField) or an optional siteCode parameter (optionalSiteParam), each
// checked as a siteCode.
const siteFlag = "site"

// addSiteFlag adds a required --site to a read whose request names the site;
// it takes the default site of the profile in effect.
func addSiteFlag(cmd *cobra.Command, site *siteCode) {
	cmd.Flags().Var(site, siteFlag, "site code, e.g. main")
	requireFlag(cmd, siteFlag)
	takeProfileDefault(cmd, siteFlag, siteFlag)
}

// profileDefault is the annotation of a flag that takes a default of the
// profile in effect (takeProfileDefault).
const profileDefault = "epx_profile_default"

// takeProfileDefault makes --flag of cmd, which cmd has just added, take the
// default of the profile in effect named setting (siteFlag or languageFlag)
// where neither the flag nor any of unless is given (applyProfileDefaults).
func takeProfileDefault(cmd *cobra.Command, flag, setting string, unless ...string) {
	annotate(cmd.Flag(flag), profileDefault, append([]string{setting}, unless...))
}

// annotate gives flag, local or persistent, the annotation key with values.
func annotate(flag *pflag.Flag, key string, values []string) {
	if flag.Annotations == nil {
		flag.Annotations = map[string][]string{}
	}
	flag.Annotations[key] = values
}

// writeScopeFlag is the flag an opt-in read requires: Emporix allows the read
// only with a write-capable scope, which epx never requests implicitly.
const writeScopeFlag = "allow-write-scope"

// addPreRunCheck makes cmd run check, once its flags are parsed and before
// it runs, after the checks added before it: an error fails the command
// before anything is sent (a usage error where check returns one). The checks
// are the PreRunE of cmd.
func addPreRunCheck(cmd *cobra.Command, check func() error) {
	before := cmd.PreRunE
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		if before != nil {
			if err := before(c, args); err != nil {
				return err
			}
		}
		return check()
	}
}

// requireOptIn makes cmd, a read with scopes, opt-in when one of them is
// write-capable (writeCapable): it adds --allow-write-scope and returns the
// check the read makes before it requests a token, which fails as a usage
// error unless the flag was given (not as false). Any other read gets no flag
// and a check that passes.
func requireOptIn(cmd *cobra.Command, scopes []string) (optedIn func() error) {
	if !writeCapable(scopes) {
		return func() error { return nil }
	}
	var allowed bool
	named := strings.Join(scopes, ", ")
	cmd.Flags().BoolVar(&allowed, writeScopeFlag, false,
		"opt in to requesting "+named+", which can also change data (required)")
	return func() error {
		if !allowed {
			return usageError{fmt.Errorf("%s requests %s, which can also change data, only when given --%s",
				cmd.CommandPath(), named, writeScopeFlag)}
		}
		return nil
	}
}
