package cli

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// cart is the part of a Cart Service cart shown in tables.
type cart struct {
	ID              string            `json:"id"`
	Status          string            `json:"status"`
	SiteCode        string            `json:"siteCode"`
	Type            string            `json:"type"`
	CustomerID      string            `json:"customerId"`
	LegalEntityID   string            `json:"legalEntityId"`
	Items           []json.RawMessage `json:"items"` // counted only
	CalculatedPrice calculatedPrice   `json:"calculatedPrice"`
	Currency        string            `json:"currency"`
	Metadata        itemMetadata      `json:"metadata"`
}

// cartItem is the part of an item of a cart shown in tables, as the items
// reads send it: not always with the unitPrice and calculatedPrice that the
// reference shows (live, Emporix sends those only for the items that the cart
// held when it was last calculated, and always on the items of the cart
// itself).
type cartItem struct {
	ID      string `json:"id"`
	Product struct {
		ID string `json:"id"`
		plainAndLocalizedName
	} `json:"product"`
	Quantity json.Number  `json:"quantity"` // kept as sent
	Price    itemPrice    `json:"price"`
	TaxLines []grossPrice `json:"itemTaxInfo"`
}

// itemPrice is the price that matched an item of a cart, per unit, as the
// price states it (net or gross).
type itemPrice struct {
	EffectiveAmount json.Number `json:"effectiveAmount"` // kept as sent
	Currency        string      `json:"currency"`
}

// cartDiscount is the part of a discount of a cart shown in tables.
type cartDiscount struct {
	Code         string      `json:"code"`
	Name         string      `json:"name"`
	Amount       json.Number `json:"amount"` // kept as sent
	Currency     string      `json:"currency"`
	DiscountRate json.Number `json:"discountRate"`
	AppliesTo    string      `json:"discountCalculationType"`
	Index        json.Number `json:"discountIndex"` // its position among the cart's discounts, which delete takes
}

// deliveryRestrictions is what limits the delivery of a cart's items.
type deliveryRestrictions struct {
	LeadTime    json.Number `json:"leadTime"` // hours, kept as sent
	NonDelivery []string    `json:"nonDelivery"`
}

// cartValidation is the result of the checks of a cart's items.
type cartValidation struct {
	IsValid *bool            `json:"isValid"`
	Items   []itemValidation `json:"itemsValidationDetails"`
}

// itemValidation is what the checks found wrong with one item of a cart.
type itemValidation struct {
	ID     string      `json:"id"`
	Errors []itemError `json:"errors"`
}

// itemError is one error that the checks found with an item of a cart.
type itemError struct {
	ErrorCode string `json:"errorCode"`
}

// String returns the item id and its error codes, e.g. "0: CODE1, CODE2".
func (v itemValidation) String() string {
	return v.ID + ": " + joinEach(v.Errors, ", ", func(e itemError) string { return e.ErrorCode })
}

// validationResult returns the result of a cart's checks, an object that
// says whether the cart is valid: one without isValid is not a result, for
// -o json as for tables.
func validationResult(body json.RawMessage) ([]json.RawMessage, error) {
	var result struct {
		IsValid *bool `json:"isValid"`
	}
	if json.Unmarshal(body, &result) != nil || result.IsValid == nil {
		return nil, errors.New("expected an object with isValid")
	}
	return []json.RawMessage{body}, nil
}

// cartScope is the scope of every cart read: Emporix allows none with a
// narrower one, so they are opt-in.
const cartScope = "cart.cart_manage"

// newCartsCommand reads and changes shopping carts, which Emporix reads only
// with the write-capable cart.cart_manage scope, so every read is opt-in.
// Emporix lists carts only by search, which shows their ids.
func newCartsCommand(a *app) *cobra.Command {
	carts := resource[cart]{
		use:   "carts",
		short: "Read and change carts and their items and discounts",
		long:  "Reads shopping carts, personal data included.",
		about: "Reads shopping carts, personal data included, with their items, discounts, delivery\n" +
			"restrictions and validation; creates, changes and deletes carts, their items and discounts.",
		path:    "/cart/{tenant}/carts",
		scopes:  []string{cartScope},
		filters: listFilters{kinds: sortable},
		search: &searchListing{
			long: "Reads the carts, personal data included, that an Emporix q filter finds (--query; without\n" +
				"it, every cart), e.g. --query 'status:OPEN siteCode:AcmeStore_NL'.",
			fields: []bodyField{queryField()},
		},
		view: view[cart]{
			singular: "cart",
			header: []string{
				columnID, columnStatus, columnSite, columnType, columnCustomer, columnLegalEntity, "ITEMS", columnTotal,
				columnModified,
			},
			legend: "STATUS is OPEN or CLOSED (sent by search only), CUSTOMER the customer id\n" +
				"(empty for an anonymous cart, whose sessionId -o json shows), ITEMS the number of items and\n" +
				"TOTAL the gross final price (items, shipping and payment fee) in the cart's currency.\n" +
				"A CLOSED cart, such as one merged into another, is listed by search but not found by get.",
			row: func(c cart, _ language) []string {
				return []string{
					c.ID, c.Status, c.SiteCode, c.Type, c.CustomerID, c.LegalEntityID, strconv.Itoa(len(c.Items)),
					c.CalculatedPrice.in(c.Currency), c.Metadata.ModifiedAt,
				}
			},
		},
	}
	reads := []*cobra.Command{
		findCartCommand(a, carts),
		newCartItemsCommand(a, carts),
		newCartDiscountsCommand(a, carts),
		childCommand(a, carts, cartDeliveryRestrictions()),
		childCommand(a, carts, cartValidationChild()),
	}
	return newSearchedResourceCommand(a, carts, slices.Concat(reads, cartWrites(a, carts))...)
}

// newCartItemsCommand returns the group of the items of a cart of carts:
// list CART_ID and get CART_ID ID, which read them, and the writes that add,
// change and remove items (cartItemWrites).
func newCartItemsCommand(a *app, carts resource[cart]) *cobra.Command {
	itemsOf := carts.nested(itemsUse)
	items := cartItems(itemsOf.use)
	return newChildGroup(a, itemsOf, items, &childGroup{
		short:  "Read and change the items of a cart",
		long:   "Reads the items of a cart; adds items to it, changes and removes them.",
		items:  []*childItem{{use: getUse, short: "Show one item of a cart", long: "Reads one item of a cart."}},
		writes: cartItemWrites(a, itemsOf, items),
	})
}

// newCartDiscountsCommand returns the group of the discounts of a cart of
// carts: list CART_ID, which reads them, and the writes that apply and
// remove discounts (cartDiscountWrites).
func newCartDiscountsCommand(a *app, carts resource[cart]) *cobra.Command {
	discountsOf := carts.nested("discounts")
	discounts := cartDiscounts(discountsOf.use)
	return newChildGroup(a, discountsOf, discounts, &childGroup{
		short:  "Read, apply and remove the discounts of a cart",
		long:   "Reads the discounts (coupons) applied to a cart; applies and removes them.",
		writes: cartDiscountWrites(a, discountsOf, discounts),
	})
}

// The flag of find that names the customer whose cart it reads, and the
// query parameter that sends it: constants, as other reads of the package use
// the same words.
const (
	cartCustomerFlag  = "customer"
	cartCustomerParam = "customerId"
)

// findCartCommand returns `find`, which reads the one cart of a site that
// criteria identify (a GET of the carts with them, which returns one cart):
// who owns it, a customer or an anonymous session, and optionally its legal
// entity and type. It never sends create, with which Emporix would create a
// missing cart.
func findCartCommand(a *app, carts resource[cart]) *cobra.Command {
	return carts.reader().command(a, &command{
		use:   "find",
		short: "Show the cart of a customer or of an anonymous session on a site",
		long: "Reads the one cart of a site that the criteria identify, personal data included: a\n" +
			"customer's (--customer) or an anonymous session's (--session). Without --legal-entity, it\n" +
			"is the cart without a legal entity, and without --type the cart without a type. epx never\n" +
			"asks Emporix to create a missing cart; when none matches, Emporix answers not found (exit 4).",
		req: request{path: carts.path, query: url.Values{siteCodeParam: {placeholderSite}}},
		fetch: &paramRead{
			params: queryFlags{
				oneOf: []paramFlag{
					{param: cartCustomerParam, flag: cartCustomerFlag, usage: "the customer `ID` whose cart to read"},
					{param: "sessionId", flag: "session", usage: "the `ID` of the anonymous session whose cart to read"},
				},
				optional: []paramFlag{
					{param: "legalEntityId", flag: "legal-entity", usage: "the legal entity `ID` of the cart"},
					{param: typeParam, flag: typeParam, usage: "the cart's `type`"},
				},
			},
			fetch: single{items: oneItem},
		},
	})
}

// cartItems declares the items of a cart, the collection sub below it, read
// with one request: the reference documents no paging for them.
func cartItems(sub string) child[cartItem] {
	return child[cartItem]{
		use:   listUse,
		sub:   sub,
		short: "List the items of a cart",
		long:  "Reads the items of a cart.",
		fetch: listing{read: (*emporix.Client).GetUnpaged},
		view: view[cartItem]{
			singular:  "cart item",
			header:    []string{columnID, columnProduct, columnName, "QUANTITY", "PRICE", "GROSS"},
			localized: true,
			legend: "PRODUCT is the product id, NAME the product's name, PRICE the price that matched the item,\n" +
				"per unit and as the price states it (net or gross), and GROSS the gross value of the item's\n" +
				"quantity before discounts, one per tax line. Emporix does not always send an item's gross\n" +
				"unit price and total (after discounts, with fees) here: only for the items that the cart\n" +
				"held when it was last calculated (by a read of the cart or a discount, for example), not\n" +
				"for those added since. The cart always has them: `epx carts get CART_ID -o json`.",
			row: func(i cartItem, lang language) []string {
				price := money{Amount: i.Price.EffectiveAmount, Currency: i.Price.Currency}
				return []string{
					i.ID, i.Product.ID, i.Product.displayName(lang), i.Quantity.String(), price.String(),
					joinEach(i.TaxLines, ", ", func(t grossPrice) string { return t.in(i.Price.Currency) }),
				}
			},
		},
	}
}

// cartDiscounts declares the discounts of a cart, the collection sub below
// it, read with one request: the reference documents no paging for them.
func cartDiscounts(sub string) child[cartDiscount] {
	return child[cartDiscount]{
		use:   listUse,
		sub:   sub,
		short: "List the discounts of a cart",
		long: "Reads the discounts (coupons) applied to a cart: AMOUNT is a fixed amount, RATE a percentage\n" +
			"of the price, APPLIES TO is SUBTOTAL or TOTAL, and INDEX its position among the cart's\n" +
			"discounts. Whether Emporix counts a discount when it calculates the cart (valid) is not\n" +
			"sent here, only with the cart's own discounts: `epx carts get CART_ID -o json`.",
		fetch: listing{read: (*emporix.Client).GetUnpaged},
		view: view[cartDiscount]{
			singular: "cart discount",
			header:   []string{columnCode, columnName, "AMOUNT", "RATE", "APPLIES TO", columnIndex},
			row: func(d cartDiscount, _ language) []string {
				amount := money{Amount: d.Amount, Currency: d.Currency}
				return []string{d.Code, d.Name, amount.String(), percent(d.DiscountRate), d.AppliesTo, d.Index.String()}
			},
		},
	}
}

// cartDeliveryRestrictions declares what limits the delivery of a cart's
// items, one object.
func cartDeliveryRestrictions() child[deliveryRestrictions] {
	return child[deliveryRestrictions]{
		use:   "delivery-restrictions",
		sub:   "dtRestrictions",
		short: "Show the lead time and non-delivery days of a cart",
		long: "Reads what limits the delivery of a cart's items: LEAD HOURS is the time needed to prepare\n" +
			"them for delivery, the longest leadTime of the items' own mixins.productCustomAttributes,\n" +
			"and NON-DELIVERY DAYS the days on which they cannot be delivered, the union of their\n" +
			"nonDelivery days, as numbers from 0 (Monday). An item gets these mixins where it is added\n" +
			"(`epx carts items create` takes them); Emporix ignores those of the item's\n" +
			"product and of its product member. The cart itself also carries leadTime and nonDelivery.",
		fetch: single{items: oneObject},
		view: view[deliveryRestrictions]{
			singular: "cart delivery restrictions",
			header:   []string{"LEAD HOURS", "NON-DELIVERY DAYS"},
			row: func(r deliveryRestrictions, _ language) []string {
				return []string{reported(r.LeadTime), strings.Join(r.NonDelivery, ", ")}
			},
		},
	}
}

// cartValidationChild declares the checks of a cart's items, which Emporix
// serves as GET and which change nothing.
func cartValidationChild() child[cartValidation] {
	return child[cartValidation]{
		use:   "validate",
		short: "Check the prices of a cart's items",
		long: "Reads whether a cart's items pass Emporix's checks, which find prices that are no longer\n" +
			"valid and prices duplicated for a product; it changes nothing. VALID is true or false and\n" +
			"ERRORS the error codes by item id (-o json adds each error's message and field).",
		fetch: single{items: validationResult},
		view: view[cartValidation]{
			singular: "cart validation",
			header:   []string{"VALID", "ERRORS"},
			long:     []string{"ERRORS"},
			row: func(v cartValidation, _ language) []string {
				return []string{formatBool(v.IsValid), joinEach(v.Items, "; ", itemValidation.String)}
			},
		},
	}
}
