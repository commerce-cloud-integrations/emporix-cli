package cli

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Table vocabulary that several services share: column headers, cell values
// and the shapes they are decoded from.

// Column headers shared by resource tables.
const (
	columnID          = "ID"
	columnCode        = "CODE"
	columnName        = "NAME"
	columnType        = "TYPE"
	columnPublished   = "PUBLISHED"
	columnVersion     = "VERSION"
	columnCurrency    = "CURRENCY"
	columnTiers       = "TIERS"
	columnFrom        = "FROM"
	columnTo          = "TO"
	columnSite        = "SITE"
	columnActive      = "ACTIVE"
	columnEnabled     = "ENABLED"
	columnModel       = "MODEL" // the model of an LLM
	columnDefault     = "DEFAULT"
	columnShipsTo     = "SHIPS TO"
	columnCountry     = "COUNTRY"
	columnStatus      = "STATUS"
	columnFailed      = "FAILED" // a count of failed items
	columnMessage     = "MESSAGE"
	columnKey         = "KEY"
	columnValue       = "VALUE"  // a long value, e.g. a setting's
	columnFields      = "FIELDS" // the names of fields, a long list
	columnStream      = "STREAM"
	columnCustomer    = "CUSTOMER"
	columnEmail       = "EMAIL"
	columnLegalEntity = "LEGAL ENTITY"
	columnGroup       = "GROUP"
	columnProduct     = "PRODUCT"
	columnParent      = "PARENT"
	columnDescription = "DESCRIPTION"
	columnProvider    = "PROVIDER"
	columnMethod      = "METHOD"
	columnZone        = "ZONE"
	columnCreated     = "CREATED"
	columnTotal       = "TOTAL"
	columnCompany     = "COMPANY"
	columnIndex       = "INDEX"
	columnModified    = "MODIFIED"
	columnItem        = "ITEM"
	columnURL         = "URL"
)

// The names of commands that several groups share: they read groups, items,
// configurations and types.
const (
	groupsUse         = "groups"
	itemsUse          = "items"
	configurationsUse = "configurations"
	typesUse          = "types"
)

// money is an amount in a currency.
type money struct {
	Amount   json.Number `json:"amount"` // kept as sent, e.g. 3.50
	Currency string      `json:"currency"`
}

// String joins the amount, its trailing zeros beyond the cents trimmed
// (trimAmount), and the currency; without an amount it is empty.
func (m money) String() string {
	if m.Amount == "" {
		return ""
	}
	return spaced(trimAmount(m.Amount.String()), m.Currency)
}

// minAmountDecimals is the number of decimals that trimAmount keeps of an
// amount at least, whatever its currency.
const minAmountDecimals = 2

// trimAmount returns a decimal amount without the trailing zeros of its
// decimals beyond the second, as tables and CSV show an amount of money (live, Emporix sends some
// amounts with 40 decimals: 35.7000…0 is shown as 35.70); 3.50 and 35.7 stay
// as sent, and so does an amount with an exponent. -o json prints it as sent.
func trimAmount(amount string) string {
	point := strings.IndexByte(amount, '.')
	if point < 0 || strings.ContainsAny(amount, "eE") {
		return amount
	}
	kept := max(len(strings.TrimRight(amount, "0")), point+1+minAmountDecimals)
	return amount[:min(kept, len(amount))] // an amount with fewer decimals stays whole
}

// valueAmount is an amount in a currency sent as {value, currency}, e.g. an
// account limit or a return total.
type valueAmount struct {
	Value    json.Number `json:"value"` // kept as sent, e.g. 1000.99
	Currency string      `json:"currency"`
}

// String joins the amount and the currency, like money.
func (v valueAmount) String() string { return money{Amount: v.Value, Currency: v.Currency}.String() }

// grossPrice is an amount of which tables show the gross value, e.g. the
// final price of an order or a tax line of a cart item.
type grossPrice struct {
	GrossValue json.Number `json:"grossValue"` // kept as sent
}

// in returns the gross value in currency, like money.
func (p grossPrice) in(currency string) string {
	return money{Amount: p.GrossValue, Currency: currency}.String()
}

// calculatedPrice is the price that Emporix calculates for an order or a
// cart, of which tables show the gross final price.
type calculatedPrice struct {
	FinalPrice grossPrice `json:"finalPrice"`
}

// in returns the gross final price in currency, like money.
func (p calculatedPrice) in(currency string) string { return p.FinalPrice.in(currency) }

// codeName is an item identified by a code with a localized name, such as a
// currency or a region.
type codeName struct {
	Code string    `json:"code"`
	Name localized `json:"name"`
}

// codeNameView is the table of code-and-name items; singular names them.
func codeNameView(singular string) view[codeName] {
	return view[codeName]{
		singular:  singular,
		header:    []string{columnCode, columnName},
		localized: true,
		row:       func(c codeName, lang language) []string { return []string{c.Code, c.Name.in(lang)} },
	}
}

// validity is the time window in which an item applies, e.g. a price, a price
// list or a segment.
type validity struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// itemMetadata is the metadata of versioned items, e.g. schemas and product
// templates, and of jobs.
type itemMetadata struct {
	Version    json.Number `json:"version"` // kept as sent; empty when missing
	CreatedAt  string      `json:"createdAt"`
	ModifiedAt string      `json:"modifiedAt"`
}

// spaced joins the parts with single spaces, each trimmed; empty parts are left out.
func spaced(parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, " ")
}

// maxValueRunes bounds a long value in a table, such as a configuration
// value: one in a column that its view declares long (view.long); CSV and
// -o json print it whole.
const maxValueRunes = 60

// cut returns text cut to at most maxValueRunes runes, ending in "…" when it
// was cut, as a table shows a long value (printer).
func cut(text string) string {
	if runes := []rune(text); len(runes) > maxValueRunes {
		return string(runes[:maxValueRunes-1]) + "…"
	}
	return text
}

// valueCell returns a JSON value for a table: text as is, other JSON
// compacted, null empty; a long value (columnValue, which tables cut).
func valueCell(value json.RawMessage) string {
	var text string
	if json.Unmarshal(value, &text) != nil {
		var compact bytes.Buffer
		if json.Compact(&compact, value) == nil {
			text = compact.String()
		}
	}
	return text
}

// joinEach formats each item and joins the results with sep.
func joinEach[T any](items []T, sep string, format func(T) string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = format(item)
	}
	return strings.Join(parts, sep)
}

// formatBool renders an optional boolean; a missing value stays empty instead
// of looking like false.
func formatBool(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

// percent renders a percentage as sent, e.g. 7%; a missing one stays empty.
func percent(value json.Number) string {
	if value == "" {
		return ""
	}
	return value.String() + "%"
}

// reported renders a number as sent, or "not reported" when it is missing:
// for the first cell of a read of one object, which identifies the row and so
// is never empty, even when Emporix has nothing to report.
func reported(value json.Number) string {
	if value == "" {
		return "not reported"
	}
	return value.String()
}

// personName is a person's first, middle and last name, as several services
// send it.
type personName struct {
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName"`
	LastName   string `json:"lastName"`
}

// fullName joins the parts of the name that are set.
func (n personName) fullName() string { return spaced(n.FirstName, n.MiddleName, n.LastName) }
