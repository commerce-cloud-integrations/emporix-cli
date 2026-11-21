package cli

import (
	"encoding/json"
	"math/big"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type rewardSummary struct {
	CustomerID   string      `json:"customerId"`
	ActivePoints json.Number `json:"activePoints"`
	Summary      *struct {
		Added    optionalList[rewardPoints] `json:"addedPointsList"`
		Redeemed []rewardPoints             `json:"redeemedPointsList"`
	} `json:"summary"`
}

func (r *rewardSummary) added() string {
	if r.Summary == nil || r.Summary.Added == nil {
		return ""
	}
	return totalPoints(*r.Summary.Added)
}

func (r *rewardSummary) redeemed() string {
	if r.Summary == nil {
		return ""
	}
	return totalPoints(r.Summary.Redeemed)
}

type optionalList[T any] *[]T

type rewardPoints struct {
	Points json.Number `json:"points"`
}

func totalPoints(entries []rewardPoints) string {
	total := new(big.Rat)
	for _, e := range entries {
		points, ok := new(big.Rat).SetString(e.Points.String())
		if !ok {
			return ""
		}
		total.Add(total, points)
	}
	if !total.IsInt() {
		return ""
	}
	return total.Num().String()
}

type redeemOption struct {
	ID     string      `json:"id"`
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Points json.Number `json:"points"`
}

const rewardCustomerPath = "/reward-points/customer/" + placeholderID

const customerSentence = "CUSTOMER is the customer id that `epx reward-points list` shows"

const redeemOptionsPath = "/reward-points/{tenant}/redeemOptions"

func newRewardPointsCommand(a *app) *cobra.Command {
	const scope = "rewardspoints.rewardpoints_view"
	const long = "Reads customers' reward points, personal data included: ACTIVE is the number of points " +
		"a customer can redeem now (expired points are left out), ADDED and REDEEMED the points added " +
		"and redeemed so far, expired ones included (a failed redemption is refunded as an " +
		"addition), so ACTIVE may be less than ADDED minus REDEEMED. Each addition and redemption is " +
		"only in `-o json`."
	summaries := reader[rewardSummary]{view: view[rewardSummary]{
		singular: "reward points summary",
		header:   []string{columnCustomer, columnActive, "ADDED", "REDEEMED"},
		row: func(r rewardSummary, _ language) []string {
			return []string{r.CustomerID, r.ActivePoints.String(), r.added(), r.redeemed()}
		},
	}, scopes: []string{scope}}
	balance := reader[json.Number]{view: view[json.Number]{
		singular: "reward points balance",
		header:   []string{"POINTS"},
		row:      func(n json.Number, _ language) []string { return []string{n.String()} },
	}, scopes: []string{scope}}
	options := reader[redeemOption]{view: view[redeemOption]{
		singular: "redeem option",
		header:   []string{columnID, columnType, columnName, "POINTS"},
		row: func(o redeemOption, _ language) []string {
			return []string{o.ID, o.Type, o.Name, o.Points.String()}
		},
	}}
	every, whole := request{path: "/reward-points/summaryBatch"}, listing{read: (*emporix.Client).GetUnpaged}
	listed := argumentIDs{columnCustomer: listedIDs(summaries, every, whole, columnCustomer)}
	optionsList := request{path: redeemOptionsPath}
	optionIDs := argumentIDs{columnID: listedIDs(options, optionsList, whole, columnID)}
	reads := []*cobra.Command{
		summaries.command(a, &command{
			use:   listUse,
			short: "List the reward points of every customer",
			long:  long,
			req:   every,
			fetch: whole,
		}),
		summaries.command(a, &command{
			use:   "get CUSTOMER",
			short: "Show the reward points of a customer",
			long:  joinSentences(long, customerSentence+"; one without reward points is not found."),
			req:   request{path: rewardCustomerPath + "/summary"},
			fetch: single{items: oneItem},
			ids:   listed,
		}),
		balance.command(a, &command{
			use:   "balance CUSTOMER",
			short: "Show the reward points balance of a customer",
			long: "Reads the reward points balance of a customer, personal data: POINTS, one number, the " +
				"points that the customer can redeem now: ACTIVE of `epx reward-points get`, which also shows " +
				"the points added and redeemed so far. " +
				customerSentence + "; an unknown one is not found (exit 4).",
			req:   request{path: rewardCustomerPath},
			fetch: single{items: oneNumber},
			ids:   listed,
		}),
		options.command(a, &command{
			use:   "redeem-options",
			short: "List the ways to redeem reward points",
			long:  "Reads the ways customers can redeem reward points: their TYPE, NAME and the POINTS they cost.",
			req:   optionsList,
			fetch: whole,
		}),
	}
	return newGroup("reward-points", "Read and change reward points and the ways to redeem them",
		long+" `epx reward-points redeem-options` lists the ways to redeem points. The writes create and "+
			"delete a customer's reward points entry (create, delete), add and redeem points (add, "+
			"redeem), and create, replace and delete the ways to redeem them (redeem-option-create, "+
			"-update, -delete).",
		append(reads, rewardPointWrites(a, listed, optionIDs)...)...)
}
