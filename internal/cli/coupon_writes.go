package cli

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// couponManage is the scope of the Coupon Service's writes of coupons and
// their redemptions.
const couponManage = "coupon.coupon_manage"

// couponRedeemOnBehalf is the scope with which a coupon is redeemed, or its
// redemption checked, for the customer that the body names.
const couponRedeemOnBehalf = "coupon.coupon_redeem_on_behalf"

// redemptionBody is the body of a coupon's redemption and of its check
// (validate): the order's total and the discount, which Emporix requires, and
// the customer, their legal entity and the order where given.
func redemptionBody() objectBody {
	return objectBody{
		moneyField("orderTotal", "order-total", "the order's total as `AMOUNT CURRENCY`, e.g. \"100 EUR\"").require(),
		moneyField("discount", "discount", "what the coupon takes off as `AMOUNT CURRENCY`").require(),
		stringField("customerNumber", "customer", "the `CUSTOMER_NUMBER` of the customer who redeems it"),
		stringField("legalEntityId", "legal-entity", "the `ID` of the customer's legal entity"),
		stringField("orderCode", "order", "the `CODE` of the order"),
	}
}

// legalEntityNote says, in the help of a redemption and of its check, how
// Emporix refuses --legal-entity (live: a customer assigned to the legal
// entity redeems with it).
const legalEntityNote = "--legal-entity names the legal entity that the customer is assigned to: Emporix\n" +
	"refuses it, with 400 Bad Request (exit 1), for a customer without a legal entity and for an\n" +
	"unknown one."

// redemptionRefusals says how Emporix refuses to redeem a coupon (live).
const redemptionRefusals = "Emporix refuses a coupon that cannot be redeemed and says why:\n" +
	"one redeemed as often as it may be with 400 Bad Request (exit 1); one that the customer may\n" +
	"not redeem, or that needs a customer where --customer is not given, with\n" +
	forbiddenRefusal + ". An unknown coupon is not found (exit 4)."

// deletedCouponNote ends the help of the writes of one coupon: Emporix keeps
// a deleted coupon, and refuses to change or delete it again (live), with a
// 403 that is a refusal (forbiddenIsRefusal).
const deletedCouponNote = "Emporix refuses a deleted coupon with " + forbiddenRefusal + "."

// couponVersionNote ends the help of the writes that send a whole or partial
// coupon.
const couponVersionNote = ifUnchangedNote + "\nWithout one, Emporix changes the coupon anyway."

// couponWrites returns the writes of coupons: create, update CODE (a PUT,
// which replaces the coupon; Emporix creates no missing one), patch CODE and
// delete CODE, the redemption of a coupon and the deletion of one, sent to
// the collection of redemptions and to one of its items, and the creation of
// a customer's referral coupon, which Emporix carries out for a token without
// scopes (live), as its reference names none.
func couponWrites(a *app, coupons resource[coupon], redemptions child[redemption]) []*cobra.Command {
	manage := []string{couponManage}
	const deleteRedemption = "redemption-delete"
	args, key := redemptions.itemArguments(deleteRedemption, []string{columnID})
	ofCoupon, profiles := argumentIDs{coupons.argument(): coupons.ids()}, customers()
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a coupon", method: http.MethodPost,
			long: "Creates the coupon in --file, a JSON object with at least its name and, for a discount,\n" +
				"discountType (PERCENT with discountPercentage, ABSOLUTE with discountAbsolute, or\n" +
				"FREE_SHIPPING), and prints {\"id\": CODE}: the code of the JSON, or one that Emporix makes\n" +
				"without one. Emporix saves codes in upper case unless the tenant configuration\n" +
				"enableCouponCodeCaseSensitivity is true. A coupon with the code exists already: a conflict\n" +
				"(exit 5), also for a deleted one, whose code stays taken.",
			example: "  epx coupons create --file coupon.json --dry-run",
			req:     request{path: coupons.path}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: coupons.itemUse("update"), short: "Replace a coupon", method: http.MethodPut,
			long: "Replaces the coupon CODE with the JSON in --file, as `epx coupons get CODE -o json`\n" +
				"prints it: a member that it leaves out is removed. `epx coupons patch` changes some\n" +
				"members only. An unknown CODE is not found (exit 4), as Emporix creates no coupon here." +
				"\n" + couponVersionNote + "\n" + deletedCouponNote,
			req: request{path: coupons.itemPath(), query: url.Values{partialParam: {queryFalse}}}, scopes: manage,
			body: &payload{}, ids: coupons.itemIDs(), answers: emporix.Answers{ForbiddenIsRefusal: true},
		}).command(a),
		(&write{
			use: coupons.itemUse("patch"), short: "Change members of a coupon", method: http.MethodPatch,
			long: "Changes the members of the coupon CODE that the JSON in --file names, such as status or\n" +
				"restrictions; the others keep their values. An unknown CODE is not found (exit 4)." +
				"\n" + couponVersionNote + "\n" + deletedCouponNote,
			req: request{path: coupons.itemPath()}, scopes: manage, body: &payload{}, ids: coupons.itemIDs(),
			answers: emporix.Answers{ForbiddenIsRefusal: true},
		}).command(a),
		(&write{
			use: coupons.itemUse("delete"), short: "Delete a coupon", method: http.MethodDelete,
			long: "Deletes the coupon CODE, which `epx coupons list` then leaves out. Emporix keeps it as\n" +
				"deleted: `epx coupons get CODE` still shows it, and its code stays taken. An unknown CODE is\n" +
				"not found (exit 4).\n" + deletedCouponNote,
			req: request{path: coupons.itemPath()}, scopes: manage, body: noBody{}, ids: coupons.itemIDs(),
			answers: emporix.Answers{ForbiddenIsRefusal: true},
		}).command(a),
		(&write{
			use: "redeem " + coupons.argument(), short: "Redeem a coupon", method: http.MethodPost,
			long: "Redeems a coupon for an order of --order-total that it takes --discount off, by\n" +
				"--customer (a customer number) where given, outside a checkout, and prints what Emporix\n" +
				"answers with: {\"id\": ID}, the redemption's id. Sent twice, it redeems the coupon twice.\n" +
				redemptionRefusals + "\n" + legalEntityNote +
				"\n`epx coupons validate` runs the same checks without redeeming it.\n" +
				"--customer needs the scope below: Emporix refuses it to a token with coupon.coupon_manage\n" +
				"alone (403 Forbidden).\n" +
				coupons.argumentSource() + ".",
			example: "  epx coupons redeem SAVE7 --order-total \"100 EUR\" --discount \"7 EUR\" --customer 10001 --dry-run",
			req:     request{path: redemptions.below(coupons.path)},
			scopes:  []string{couponRedeemOnBehalf}, body: &payload{typed: redemptionBody()}, ids: ofCoupon,
			answers: emporix.Answers{ForbiddenIsRefusal: true},
		}).command(a),
		(&write{
			use:   strings.Join(append([]string{deleteRedemption, coupons.argument()}, args...), " "),
			short: "Delete a redemption of a coupon", method: http.MethodDelete,
			long: "Deletes the redemption ID of a coupon, which then counts one redemption less: one that was\n" +
				"redeemed as often as it may be can be redeemed again.\n" + coupons.argumentSentence(notFound) + "\n" +
				itemSentence(args, []string{columnID}, childListLine(coupons, redemptions)),
			req: request{path: redemptions.below(coupons.path) + "/" + key}, scopes: manage, body: noBody{},
			ids: ofCoupon,
		}).command(a),
		(&write{
			use: "referral-create " + profiles.argument(), short: "Create the referral coupon of a customer",
			method: http.MethodPost,
			long: "Creates the referral coupon of a customer, from the referral settings of the main site,\n" +
				"and prints it; a customer has one referral coupon at most, so where it exists, Emporix\n" +
				"answers with it and creates none. A deleted referral coupon stays the customer's: Emporix\n" +
				"answers with it and makes no new one (`epx coupons referral` shows it as VALID). Emporix\n" +
				"keeps the referral settings in the site's couponSettings mixin; where the main site has\n" +
				"none, it answers 400 Bad Request (exit 1). CUSTOMER_NUMBER is the NUMBER that\n" +
				"`epx customers list` shows; an unknown one is not found (exit 4).",
			req: request{path: referralCouponsPath + "/" + placeholderID}, body: noBody{}, readScoped: true,
			ids: argumentIDs{profiles.argument(): profiles.ids()},
		}).command(a),
	}
}
