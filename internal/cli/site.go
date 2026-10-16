package cli

import (
	"fmt"
	"strings"
)

// siteCodeParam is the query parameter that filters a read by site.
const siteCodeParam = "siteCode"

// siteCode is a --site value; it implements pflag.Value, so codes
// that cannot address one site are rejected while flags are parsed.
type siteCode string

// Set implements pflag.Value.
func (s *siteCode) Set(value string) error {
	if !isPathSegment(value) {
		return fmt.Errorf("invalid site %q", value)
	}
	*s = siteCode(value)
	return nil
}

func (s *siteCode) String() string { return string(*s) }

// Type implements pflag.Value.
func (s *siteCode) Type() string { return "site" }

// siteCodes is a repeatable --site value, the sites in their order; it
// implements pflag.Value, each code checked as a siteCode.
type siteCodes []string

// Set implements pflag.Value: it adds one site.
func (s *siteCodes) Set(value string) error {
	var site siteCode
	if err := site.Set(value); err != nil {
		return err
	}
	*s = append(*s, value)
	return nil
}

func (s *siteCodes) String() string { return strings.Join(*s, ",") }

// Type implements pflag.Value: each value is a siteCode.
func (s *siteCodes) Type() string { return new(siteCode).Type() }

// optionalSiteParam is an optional --site that sets the siteCode query
// parameter of a request, with usage as its help; its value is checked as a
// siteCode while flags are parsed.
func optionalSiteParam(usage string) paramFlag {
	return paramFlag{param: siteCodeParam, flag: siteFlag, usage: usage,
		check: func(value string) error { return new(siteCode).Set(value) }}
}

// settingsSiteParam is an optionalSiteParam that selects the site whose
// settings apply, which Emporix takes as main without it: it takes the
// default site of the profile in effect where it is not given, as a
// required site does (addSiteFlag).
func settingsSiteParam(usage string) paramFlag {
	site := optionalSiteParam(usage)
	site.profileSetting = siteFlag
	return site
}
