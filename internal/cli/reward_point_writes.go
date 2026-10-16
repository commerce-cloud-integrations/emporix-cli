package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// rewardPointsManage is the scope of the Reward Points Service's writes:
// customers' reward points and the ways to redeem them.
const rewardPointsManage = "rewardspoints.rewardpoints_manage"

// redeemConflict is what a 409 Conflict of a redemption of reward points
// means (live): the customer has fewer points active than it redeems.
const redeemConflict = "the customer has fewer active points than it redeems: `epx reward-points balance` " +
	"shows them"

// rewardSite is the optional --site of the writes that add reward points:
// the site whose reward points settings apply, sent as siteCode, which takes
// the default site of the profile in effect. Without either, Emporix applies
// those of site main, as its reference documents and a live create showed.
func rewardSite() queryFlags {
	return queryFlags{optional: []paramFlag{
		settingsSiteParam("the site `CODE` whose reward points settings apply, sent as siteCode"),
	}}
}

// rewardPointWrites returns the writes of customers' reward points: create
// and delete CUSTOMER, which create and delete the customer's entry, add and
// redeem CUSTOMER, which change the points, and the creation, update and
// deletion of a way to redeem points. The reference documents that deletion
// for a customer's token only; live, a service token with the manage scope
// deletes the option (204) and one without it is refused (403). Points are
// counts (countField), as the reference declares them integers. Create and
// add take the site whose reward points settings apply (rewardSite). Create
// repeats CUSTOMER as customerId in its body, which Emporix requires (live:
// 500 without it). CUSTOMER completes from customers, the ID of a redeem
// option from options.
func rewardPointWrites(a *app, customers, options argumentIDs) []*cobra.Command {
	manage := []string{rewardPointsManage}
	const customer = "\n" + customerSentence + "."
	const site = "The reward points settings of the site --site apply, the default site of the profile in\n" +
		"effect without it, and site main without either: how long the points stay valid. Emporix\n" +
		"keeps them in the site's couponSettings mixin; it accepts a site without them, also an\n" +
		"unknown site code, and the points then stay valid 36 months."
	const noEntry = "\nA customer without a reward points entry is not found (exit 4)."
	const optionID = "the redeem option ID, as `epx reward-points redeem-options` shows it"
	option := request{path: redeemOptionsPath + "/" + placeholderID}
	optionDelete := deleteWrite("redeem-option-delete "+columnID, "Delete a way to redeem reward points",
		"Deletes "+optionID+"; it sends no body."+unknownAsDeleted(columnID), option, manage)
	optionDelete.ids = options
	return []*cobra.Command{
		(&write{
			use: "create " + columnCustomer, short: "Create the reward points entry of a customer",
			method: http.MethodPost,
			long: "Creates the reward points entry of a customer, with the points it starts with where\n" +
				"--points gives them. Emporix answers without a body, so epx prints nothing. A customer\n" +
				"with an entry is a conflict (exit 5). Emporix does not check the customer: an unknown one\n" +
				"gets an entry too.\n" +
				"`epx reward-points add` adds points to an entry.\n" + site + "\n" +
				"The JSON of --file must give customerId, CUSTOMER again (Emporix answers 400 Bad Request to\n" +
				"another and 500 Internal Server Error without one, exit 1, and creates no entry), and the\n" +
				"points it starts with as startPoints, e.g.\n" +
				"{\"customerId\": \"C1001\", \"startPoints\": {\"points\": 20}}." + customer,
			example: "  epx reward-points create C1001 --site AcmeStore_NL --points 20 --description \"Welcome\"",
			req:     request{path: rewardCustomerPath}, params: rewardSite(),
			scopes: manage, ids: customers,
			body: &payload{typed: objectBody{
				argumentField("customerId", columnCustomer),
				countField("startPoints.points", "points", "the number of `POINTS` the entry starts with"),
				stringField("startPoints.description", "description", "what the points it starts with are for").
					onlyWith("points"),
			}},
		}).command(a),
		(&write{
			use: "delete " + columnCustomer, short: "Delete the reward points of a customer", method: http.MethodDelete,
			long: "Deletes the reward points entry of a customer, with the points added and redeemed so far.\n" +
				"Emporix answers 200 OK, also for an unknown customer or one without an entry: neither is an\n" +
				"error." + customer,
			req:    request{path: rewardCustomerPath},
			scopes: manage, body: noBody{}, ids: customers,
		}).command(a),
		(&write{
			use: "add " + columnCustomer, short: "Add reward points to a customer", method: http.MethodPost,
			long: "Adds --points reward points to a customer. Emporix answers without a body, so epx prints\n" +
				"nothing. Sent twice, it adds them twice.\n" + site +
				"\n--valid-until, which gives these points their own end of validity, takes a time such as\n" +
				"2027-12-31T23:59:59Z: unlike the redeem options, Emporix reads a time that ends in Z." +
				noEntry + customer,
			example: "  epx reward-points add C1001 --site AcmeStore_NL --points 50 --description \"Goodwill\"",
			req:     request{path: rewardCustomerPath + "/addPoints"}, params: rewardSite(),
			scopes: manage, ids: customers,
			body: &payload{typed: objectBody{
				countField("points", "points", "the number of `POINTS` to add").require(),
				stringField("description", "description", "what the points are for"),
				stringField("reference-id", "order", "the `ID` of the order that earned the points"),
				stringField("validUntil", "valid-until", "the `TIME` until which the points can be redeemed, "+
					"e.g. 2027-12-31T23:59:59Z"),
			}},
		}).command(a),
		(&write{
			use: "redeem " + columnCustomer, short: "Redeem reward points of a customer", method: http.MethodPost,
			long: "Redeems --points reward points of a customer without making a coupon of them, the oldest\n" +
				"points that are still valid first. Emporix answers without a body, so epx prints nothing.\n" +
				"More points than are active (ACTIVE of `epx reward-points get`) are a conflict (exit 5).\n" +
				"Sent twice, it redeems them twice." + noEntry + customer,
			example: "  epx reward-points redeem C1001 --points 50 --description \"Gift wrapping\"",
			req:     request{path: rewardCustomerPath + "/redeemPoints"}, scopes: manage, ids: customers,
			body: &payload{typed: objectBody{
				countField("points", "points", "the number of `POINTS` to redeem").require(),
				stringField("description", "description", "what the points are redeemed for"),
			}},
			conflict: redeemConflict,
		}).command(a),
		(&write{
			use: "redeem-option-create", short: "Create a way to redeem reward points", method: http.MethodPost,
			long: "Creates a way to redeem reward points from the JSON in --file: its type (coupon), name,\n" +
				"the points it costs and, for a coupon, the coupon that it makes of them, as\n" +
				"`epx coupons create` takes it. It prints the option that Emporix answers with; sending the\n" +
				"same JSON twice creates two. The coupon's restrictions.validFrom and validUntil take epoch\n" +
				"milliseconds or a time with a numeric offset, e.g. 2027-01-01T00:00:00.000+00:00: Emporix\n" +
				"refuses a time that ends in Z (400 Bad Request).",
			req: request{path: redeemOptionsPath}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: "redeem-option-update " + columnID, short: "Replace a way to redeem reward points",
			method: http.MethodPut,
			long: "Replaces " + optionID + ", with the\n" +
				"JSON in --file, as redeem-option-create takes it.\n" + ifUnchangedNote,
			req: option, scopes: manage, body: &payload{}, ids: options,
		}).command(a),
		optionDelete.command(a),
	}
}
