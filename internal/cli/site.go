package cli

import (
	"fmt"
	"strings"
)

const siteCodeParam = "siteCode"

type siteCode string

func (s *siteCode) Set(value string) error {
	if !isPathSegment(value) {
		return fmt.Errorf("invalid site %q", value)
	}
	*s = siteCode(value)
	return nil
}

func (s *siteCode) String() string { return string(*s) }

func (s *siteCode) Type() string { return "site" }

type siteCodes []string

func (s *siteCodes) Set(value string) error {
	var site siteCode
	if err := site.Set(value); err != nil {
		return err
	}
	*s = append(*s, value)
	return nil
}

func (s *siteCodes) String() string { return strings.Join(*s, ",") }

func (s *siteCodes) Type() string { return new(siteCode).Type() }

func optionalSiteParam(usage string) paramFlag {
	return paramFlag{param: siteCodeParam, flag: siteFlag, usage: usage,
		check: func(value string) error { return new(siteCode).Set(value) }}
}

func settingsSiteParam(usage string) paramFlag {
	site := optionalSiteParam(usage)
	site.profileSetting = siteFlag
	return site
}
