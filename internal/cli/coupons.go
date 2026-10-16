package cli

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// coupon is the part of a Coupon Service coupon shown in tables.
type coupon struct {
	Code               string      `json:"code"`
	Name               string      `json:"name"`
	DiscountType       string      `json:"discountType"`
	DiscountPercentage json.Number `json:"discountPercentage"` // kept as sent
	DiscountAbsolute   money       `json:"discountAbsolute"`
	Status             string      `json:"status"`
	RedemptionCount    json.Number `json:"redemptionCount"`
	MaxRedemptions     json.Number `json:"maxRedemptions"`
	Restrictions       struct {
		ValidUntil string `json:"validUntil"`
	} `json:"restrictions"`
}

// discount returns the coupon's discount: a percentage, an amount or free
// shipping.
func (c *coupon) discount() string {
	switch c.DiscountType {
	case "PERCENT":
		return percent(c.DiscountPercentage)
	case "ABSOLUTE":
		return c.DiscountAbsolute.String()
	default:
		return c.DiscountType
	}
}

// redeemed returns how often the coupon was redeemed, out of its maximum when
// it has one (-1, the default, means unlimited).
func (c *coupon) redeemed() string {
	if limit, err := c.MaxRedemptions.Float64(); err != nil || limit < 0 {
		return c.RedemptionCount.String()
	}
	return c.RedemptionCount.String() + "/" + c.MaxRedemptions.String()
}

// redemption is the part of a coupon redemption shown in tables.
type redemption struct {
	ID             string `json:"id"`
	OrderCode      string `json:"orderCode"`
	CustomerNumber string `json:"customerNumber"`
	LegalEntityID  string `json:"legalEntityId"`
	OrderTotal     money  `json:"orderTotal"`
	Discount       money  `json:"discount"`
	RedeemedAt     string `json:"redeemedAt"`
}

// columnDiscount is the discount of a coupon, and the one a redemption granted.
const columnDiscount = "DISCOUNT"

// newCouponsCommand reads coupons and their redemptions, which Emporix reads
// only with the write-capable coupon.coupon_manage scope, so those reads are
// opt-in, and creates, changes, redeems and deletes them (couponWrites).
func newCouponsCommand(a *app) *cobra.Command {
	const long = "Reads coupons, customer ids included (issuedTo, validFor in -o json). DISCOUNT is a\n" +
		"percentage, an amount or FREE_SHIPPING. STATUS is VALID, INACTIVE, EXPIRED or USED, as\n" +
		"Emporix computes it: --query status:VALUE matches no coupon and --sort status is ignored.\n" +
		"REDEEMED counts redemptions, out of the maximum when there is one; Emporix counts those for\n" +
		"a legal entity apart, per legal entity (legalEntityRedemptions, only in -o json). Other\n" +
		"restrictions (valid from, minimum order value, customers, categories, segments) are only in\n" +
		"-o json.\n" +
		"`epx coupons redemptions` and `redemption` read the redemptions of a coupon, all or one.\n" +
		"`epx coupons list` leaves out deleted coupons; `epx coupons get CODE` still shows one, in\n" +
		"the table as a live one (-o json has deleted: true). Deleted coupons are listed too, one\n" +
		"page at a time, by\n" +
		"`epx api get /coupon/{tenant}/coupons -p showDeleted=true --scope coupon.coupon_read`."
	coupons := resource[coupon]{
		use:   "coupons",
		short: "Read, create, change and redeem coupons",
		long:  long,
		about: long + "\n`epx coupons create`, `update`, `patch` and `delete` change coupons, `redeem` redeems one\n" +
			"outside a checkout, `redemption-delete` deletes a redemption, and `referral-create` creates\n" +
			"the referral coupon of a customer.",
		path:    "/coupon/{tenant}/coupons",
		scopes:  []string{"coupon.coupon_read"},
		filters: listFilters{kinds: filterable},
		read:    (*emporix.Client).GetAllHybrisCounted, // the total comes only as hybris-count
		view:    couponView(),
	}
	redemptions := child[redemption]{
		use:   "redemptions",
		short: "List the redemptions of a coupon",
		long: "Reads the redemptions of a coupon (personal data): the order, customer and legal entity\n" +
			"of each, and when it was redeemed (`epx coupons redemption` shows one).",
		scopes: []string{couponManage},
		fetch:  listing{filters: listFilters{kinds: filterable}},
		view: view[redemption]{
			singular: "redemption",
			header: []string{
				columnID, "ORDER", columnCustomer, columnLegalEntity, "ORDER TOTAL", columnDiscount, "REDEEMED AT",
			},
			legend: "ORDER TOTAL is the order's value and DISCOUNT the amount the coupon took off it. The\n" +
				"narrower coupon.coupon_redeem scope reads only the redemptions a caller made itself.",
			row: func(r redemption, _ language) []string {
				return []string{r.ID, r.OrderCode, r.CustomerNumber, r.LegalEntityID, r.OrderTotal.String(),
					r.Discount.String(), r.RedeemedAt}
			},
		},
	}
	reads := []*cobra.Command{
		childCommand(a, coupons, redemptions),
		childItemCommand(a, coupons, redemptions, &childItem{
			use:   "redemption",
			short: "Show one redemption of a coupon",
			long:  "Reads one redemption of a coupon (personal data).",
		}),
		newReferralCouponCommand(a), newCouponValidationCommand(a, coupons),
	}
	return newResourceCommand(a, coupons, append(reads, couponWrites(a, coupons, redemptions)...)...)
}

// couponView is the table of coupons, also of a customer's referral coupon.
func couponView() view[coupon] {
	return view[coupon]{
		singular: "coupon",
		header:   []string{columnCode, columnName, columnDiscount, columnStatus, "REDEEMED", "VALID UNTIL"},
		row: func(c coupon, _ language) []string {
			return []string{c.Code, c.Name, c.discount(), c.Status, c.redeemed(), c.Restrictions.ValidUntil}
		},
	}
}

// referralCouponsPath is the collection of customers' referral coupons, each
// below the customer's number.
const referralCouponsPath = "/coupon/{tenant}/referral-coupons"

// newReferralCouponCommand reads the referral coupon of a customer. The
// reference declares no scope for it, and Emporix answers an unknown customer
// to a token without scopes.
func newReferralCouponCommand(a *app) *cobra.Command {
	profiles := customers()
	return reader[coupon]{view: couponView()}.command(a, &command{
		use:   "referral " + profiles.argument(),
		short: "Show the referral coupon of a customer",
		long: "Reads the referral coupon of a customer: the coupon issued to the customer (issuedTo in\n" +
			"-o json) to share with the people they refer, also a deleted one, which the table shows as\n" +
			"a live one (-o json has deleted: true). Its columns are those of `epx coupons list`.\n" +
			profiles.argumentSource() + "; a customer without a\n" +
			"referral coupon, or an unknown one, is not found (exit 4).",
		req:   request{path: referralCouponsPath + "/" + placeholderID},
		fetch: single{items: oneItem},
		ids:   argumentIDs{profiles.argument(): profiles.ids()},
	})
}

// newCouponValidationCommand returns `validate COUPON_CODE`, whether a coupon
// can be redeemed: a check that Emporix serves as POST without redeeming it,
// with the write-capable coupon.coupon_redeem_on_behalf scope, so it is
// opt-in. The reference names only that scope, required to name the customer;
// whether Emporix validates without it otherwise is not documented, so it is
// requested also without --customer, as a reader's scopes do not depend on
// its flags.
func newCouponValidationCommand(a *app, coupons resource[coupon]) *cobra.Command {
	check := reader[bool]{view: answerView("coupon validation", "REDEEMABLE"),
		scopes: []string{couponRedeemOnBehalf}}
	return check.command(a, &command{
		use:   "validate " + coupons.argument(),
		short: "Check whether a coupon can be redeemed",
		long: "Checks whether a coupon can be redeemed for an order of --order-total that it takes\n" +
			"--discount off, by --customer (a customer number) where given: Emporix runs every check of\n" +
			"a redemption without redeeming the coupon. REDEEMABLE is true when it can be. Emporix\n" +
			"refuses a coupon that cannot be redeemed with an error that says why: one that the customer\n" +
			"may not redeem with " + forbiddenRefusal + ". An unknown coupon is not found (exit 4).\n" +
			legalEntityNote + "\n" +
			"Emporix answers a redeemable coupon without a body; -o json prints true then. The scope,\n" +
			"which epx requests also without --customer, lets a credential redeem coupons on a\n" +
			"customer's behalf too.\n" +
			coupons.argumentSource() + ".",
		req:   request{path: coupons.path + "/" + placeholderID + "/validation"},
		ids:   argumentIDs{coupons.argument(): coupons.ids()},
		fetch: bodyRead{method: http.MethodPost, body: redemptionBody(), fetch: statusCheck(), forbiddenIsRefusal: true},
	})
}
