package cli

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

const mixinSiteFlag = "mixin-site"

// The Product and Schema Services merge the overlays of the site this header
// names into the site-aware mixins of what they answer.
const siteHeader = "emporix-site"

const mixinSiteHelp = "Where a mixin's schema is site-aware, an item may hold, besides the mixin NAME, an overlay " +
	"for a site, NAME__SITE, with the values that differ there. --" + mixinSiteFlag + " SITE sends SITE as " +
	"the " + siteHeader + " header: Emporix then merges the overlay of SITE into NAME (nested objects " +
	"member by member; arrays and other values are replaced, not merged; a null leaves the " +
	"base value) and leaves out every overlay. A site without an overlay, or one the tenant " +
	"lacks (codes are case-sensitive; epx does not check them), gets NAME as stored. Without " +
	"--" + mixinSiteFlag + ", mixins and overlays are read as stored, whatever the default site of the " +
	"profile. Tables leave mixins (attribute values) out; `-o json` and `-o yaml` print them, as " +
	"--" + streamFlag + " does on a list.\n\n" +
	"A merged item is that site's view, not the item as stored. Do not write it back with " +
	"update, bulk-update or (products) import, which replace whole items and would delete every " +
	"overlay, nor build a patch from it: NAME would take the site's values."

type mixinSiteRead struct{ fetch fetcher }

func (m mixinSiteRead) bind(cmd *cobra.Command) (readFunc, shape) {
	site := &paramValue{paramFlag: paramFlag{param: siteHeader, flag: mixinSiteFlag,
		usage: "merge the site-aware mixin overlays of the site `SITE` (NAME__SITE) into their mixins",
		check: mixinSite}}
	site.register(cmd)
	read, items := m.fetch.bind(cmd)
	return func(c *emporix.Client, ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
		if site.given == "" {
			return read(c, ctx, path, query)
		}
		merging, err := c.WithHeader(site.param, site.given)
		if err != nil {
			return nil, err
		}
		return read(merging, ctx, path, query)
	}, items
}

func mixinSite(value string) error {
	switch trimmed := strings.TrimSpace(value); {
	case trimmed == "":
		return errors.New("names no site")
	case trimmed != value:
		return errors.New("has spaces around the site code")
	}
	return emporix.CheckHeaderValue(value)
}

func (v view[T]) withMixinSite(fetch fetcher) fetcher {
	if !v.siteAware {
		return fetch
	}
	return mixinSiteRead{fetch: fetch}
}
