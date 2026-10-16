package cli

// couponReads returns the Coupon Service reads.
func couponReads() []readCase {
	return []readCase{
		{
			// DISCOUNT follows the discount type (a percentage may be missing); REDEEMED shows the
			// maximum when there is one (-1 is unlimited). Coupons send their total only as
			// hybris-count, when asked with totalCount=true.
			args:    []string{"coupons", "list"},
			path:    "/coupon/acme1/coupons",
			mode:    getAllHybrisCounted,
			scopes:  []string{"coupon.coupon_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"code":"SAVE7","name":"Save 7%","discountType":"PERCENT","discountPercentage":7,
				"status":"VALID","redemptionCount":3,"maxRedemptions":100,
				"restrictions":{"validUntil":"2026-12-31T23:59:59Z","minOrderValue":{"amount":150,"currency":"EUR"}}},
				{"code":"TEN","name":"Ten off","discountType":"ABSOLUTE","discountAbsolute":{"amount":10.00,"currency":"EUR"},
				"status":"USED","redemptionCount":1,"maxRedemptions":1},
				{"code":"SHIP","discountType":"FREE_SHIPPING","status":"INACTIVE","redemptionCount":0},
				{"code":"ANY","discountType":"PERCENT","status":"VALID","redemptionCount":4,"maxRedemptions":-1}]`,
			want: "CODE   NAME     DISCOUNT       STATUS    REDEEMED  VALID UNTIL\n" +
				"SAVE7  Save 7%  7%             VALID     3/100     2026-12-31T23:59:59Z\n" +
				"TEN    Ten off  10.00 EUR      USED      1/1       \n" +
				"SHIP            FREE_SHIPPING  INACTIVE  0         \n" +
				"ANY                            VALID     4         \n",
		},
		{
			args:   []string{"coupons", "get", "SAVE 7"},
			path:   "/coupon/acme1/coupons/SAVE%207",
			scopes: []string{"coupon.coupon_read"},
			mode:   getWhole,
			body:   `{"code":"SAVE 7","name":"Save 7%","discountType":"PERCENT","discountPercentage":7.5,"status":"VALID"}`,
			want: "CODE    NAME     DISCOUNT  STATUS  REDEEMED  VALID UNTIL\n" +
				"SAVE 7  Save 7%  7.5%      VALID             \n",
		},
		{
			// Opt-in: the write-capable scope is requested only with --allow-write-scope.
			args:    []string{"coupons", "redemptions", "SAVE 7", "--allow-write-scope"},
			path:    "/coupon/acme1/coupons/SAVE%207/redemptions",
			scopes:  []string{"coupon.coupon_manage"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"r1","code":"SAVE 7","orderCode":"O-100","customerNumber":"C0123",
				"orderTotal":{"amount":299.30,"currency":"EUR"},"discount":{"amount":20.95,"currency":"EUR"},
				"redeemedAt":"2026-04-02T17:18:53.701Z"},
				{"id":"r2","customerNumber":"C0456","legalEntityId":"le1","orderTotal":{"amount":80,"currency":"EUR"},
				"discount":{"amount":5.6,"currency":"EUR"}}]`,
			want: "ID  ORDER  CUSTOMER  LEGAL ENTITY  ORDER TOTAL  DISCOUNT   REDEEMED AT\n" +
				"r1  O-100  C0123                   299.30 EUR   20.95 EUR  2026-04-02T17:18:53.701Z\n" +
				"r2         C0456     le1           80 EUR       5.6 EUR    \n",
		},
		{
			args:   []string{"coupons", "redemption", "SAVE7", "r/1", "--allow-write-scope"},
			path:   "/coupon/acme1/coupons/SAVE7/redemptions/r%2F1",
			scopes: []string{"coupon.coupon_manage"},
			mode:   getWhole,
			body: `{"id":"r/1","code":"SAVE7","orderCode":"O-100","customerNumber":"C0123",
				"orderTotal":{"amount":100,"currency":"EUR"},"discount":{"amount":7,"currency":"EUR"},
				"redeemedAt":"2026-04-02T17:18:53.701Z"}`,
			want: "ID   ORDER  CUSTOMER  LEGAL ENTITY  ORDER TOTAL  DISCOUNT  REDEEMED AT\n" +
				"r/1  O-100  C0123                   100 EUR      7 EUR     2026-04-02T17:18:53.701Z\n",
		},
		{
			// A referral coupon is a coupon issued to the customer; the reference declares no scope.
			args: []string{"coupons", "referral", "100/01"},
			path: "/coupon/acme1/referral-coupons/100%2F01",
			mode: getWhole,
			body: `{"code":"REF_30","name":"Referral","discountType":"ABSOLUTE","discountAbsolute":{"amount":8,"currency":"EUR"},
				"issuedTo":"100/01","referralCoupon":true,"maxRedemptions":-1,"redemptionCount":2,"status":"VALID"}`,
			want: "CODE    NAME      DISCOUNT  STATUS  REDEEMED  VALID UNTIL\n" +
				"REF_30  Referral  8 EUR     VALID   2         \n",
		},
	}
}

// rewardPointReads returns the Reward Points Service reads.
func rewardPointReads() []readCase {
	return []readCase{
		{
			// ADDED and REDEEMED sum the points exactly (5.0 and 5e0 are whole; a sum that is not
			// whole stays empty); a summary without additions leaves ADDED unknown.
			args:   []string{"reward-points", "list"},
			path:   "/reward-points/summaryBatch",
			scopes: []string{"rewardspoints.rewardpoints_view"},
			mode:   getWhole,
			body: `[{"customerId":"c1","activePoints":150,"summary":{"addedPointsList":[{"points":100},{"points":80}],
				"redeemedPointsList":[{"points":30}]}},{"customerId":"c2","activePoints":0,
				"summary":{"addedPointsList":[{"points":1.5}]}},{"customerId":"c3","activePoints":10,
				"summary":{"addedPointsList":[{"points":5.0},{"points":5e0},{"points":9223372036854775807},{"points":1}],
				"redeemedPointsList":[]}},{"customerId":"c4","activePoints":150},
				{"customerId":"c5","summary":{"redeemedPointsList":[{"points":3}]}}]`,
			want: "CUSTOMER  ACTIVE  ADDED                REDEEMED\n" +
				"c1        150     180                  30\n" +
				"c2        0                            0\n" +
				"c3        10      9223372036854775818  0\n" +
				"c4        150                          \n" +
				"c5                                     3\n",
		},
		{
			args:   []string{"reward-points", "get", "c/1"},
			path:   "/reward-points/customer/c%2F1/summary",
			scopes: []string{"rewardspoints.rewardpoints_view"},
			mode:   getWhole,
			body:   `{"customerId":"c/1","activePoints":5,"summary":{"addedPointsList":[{"points":5}]}}`,
			want:   "CUSTOMER  ACTIVE  ADDED  REDEEMED\nc/1       5       5      0\n",
		},
		{
			// The balance is one number, shown as sent.
			args:   []string{"reward-points", "balance", "c/1"},
			path:   "/reward-points/customer/c%2F1",
			scopes: []string{"rewardspoints.rewardpoints_view"},
			mode:   getWhole,
			body:   ` 9223372036854775808 `,
			want:   "POINTS\n9223372036854775808\n",
		},
		{
			args: []string{"reward-points", "redeem-options"},
			path: "/reward-points/acme1/redeemOptions",
			mode: getWhole,
			body: `[{"id":"o1","type":"coupon","name":"5 EUR off","points":500,"coupon":{"code":"RP5","name":"5 off",
				"discountType":"ABSOLUTE","discountAbsolute":{"amount":5,"currency":"EUR"}}}]`,
			want: "ID  TYPE    NAME       POINTS\no1  coupon  5 EUR off  500\n",
		},
	}
}
