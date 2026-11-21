package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func requireFlag(cmd *cobra.Command, flag string) {
	// MarkFlagRequired fails only for an undefined flag, and the flag was just added.
	_ = cmd.MarkFlagRequired(flag)
	noteRequired(cmd, flag, "required")
}

func requireFlagUnless(cmd *cobra.Command, flag, other string) {
	cmd.MarkFlagsOneRequired(other, flag)
	noteRequired(cmd, flag, "required without --"+other)
}

func noteRequired(cmd *cobra.Command, flag, note string) {
	given := cmd.Flags().Lookup(flag)
	if usage, closed := strings.CutSuffix(given.Usage, ")"); closed {
		given.Usage = usage + "; " + note + ")"
	} else {
		given.Usage += " (" + note + ")"
	}
}

const languageFlag = "language"

func addLanguageFlag(cmd *cobra.Command, lang *language) {
	cmd.Flags().Var(lang, languageFlag,
		"language of localized table text; a missing translation shows another, labeled (-o json prints all)")
	takeProfileDefault(cmd, languageFlag, languageFlag)
}

const siteFlag = "site"

func addSiteFlag(cmd *cobra.Command, site *siteCode) {
	cmd.Flags().Var(site, siteFlag, "site code, e.g. main")
	requireFlag(cmd, siteFlag)
	takeProfileDefault(cmd, siteFlag, siteFlag)
}

const profileDefault = "epx_profile_default"

func takeProfileDefault(cmd *cobra.Command, flag, setting string, unless ...string) {
	annotate(cmd.Flag(flag), profileDefault, append([]string{setting}, unless...))
}

func annotate(flag *pflag.Flag, key string, values []string) {
	if flag.Annotations == nil {
		flag.Annotations = map[string][]string{}
	}
	flag.Annotations[key] = values
}

const writeScopeFlag = "allow-write-scope"

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
