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

// mixinSiteFlag is the flag of a read of items with site-aware mixins (a
// siteAware view) that names the site whose overlays Emporix merges into
// them (mixinSiteRead). It is not --site: a profile's default site never
// sets it, so that a read without it sends no emporix-site and Emporix
// answers the mixins as stored.
const mixinSiteFlag = "mixin-site"

// siteHeader is the header that names the site whose overlays the Product
// and Schema Services merge into the site-aware mixins of what they answer.
const siteHeader = "emporix-site"

// mixinSiteHelp is the help of --mixin-site that follows the legend of a
// read of items with site-aware mixins (reader.help).
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

// mixinSiteRead reads as fetch reads, with --mixin-site SITE, which sends
// SITE as the emporix-site header (emporix.Client.WithHeader); without it,
// it sends what fetch sends, nothing more. Its help is mixinSiteHelp.
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

// mixinSite refuses a value of --mixin-site that is blank, which names no
// site, that has spaces around the site code, which a header sent over
// HTTP/2 keeps, or that cannot be sent as a header value
// (emporix.CheckHeaderValue, which Client.WithHeader checks again before it
// sends it).
func mixinSite(value string) error {
	switch trimmed := strings.TrimSpace(value); {
	case trimmed == "":
		return errors.New("names no site")
	case trimmed != value:
		return errors.New("has spaces around the site code")
	}
	return emporix.CheckHeaderValue(value)
}

// withMixinSite returns how a read of v's items reads as fetch reads: with
// --mixin-site where they may hold site-aware mixins (siteAware), else as
// fetch alone.
func (v view[T]) withMixinSite(fetch fetcher) fetcher {
	if !v.siteAware {
		return fetch
	}
	return mixinSiteRead{fetch: fetch}
}
