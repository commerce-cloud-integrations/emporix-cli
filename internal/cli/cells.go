package cli

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

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
	columnModel       = "MODEL"
	columnDefault     = "DEFAULT"
	columnShipsTo     = "SHIPS TO"
	columnCountry     = "COUNTRY"
	columnStatus      = "STATUS"
	columnFailed      = "FAILED"
	columnRead        = "READ"
	columnMessage     = "MESSAGE"
	columnKey         = "KEY"
	columnValue       = "VALUE"
	columnFields      = "FIELDS"
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

const (
	groupsUse         = "groups"
	itemsUse          = "items"
	configurationsUse = "configurations"
	typesUse          = "types"
)

type money struct {
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
}

func (m money) String() string {
	if m.Amount == "" {
		return ""
	}
	return spaced(trimAmount(m.Amount.String()), m.Currency)
}

const minAmountDecimals = 2

// Emporix sends some amounts with 40 decimals (35.7000…0).
func trimAmount(amount string) string {
	point := strings.IndexByte(amount, '.')
	if point < 0 || strings.ContainsAny(amount, "eE") {
		return amount
	}
	kept := max(len(strings.TrimRight(amount, "0")), point+1+minAmountDecimals)
	return amount[:min(kept, len(amount))]
}

type valueAmount struct {
	Value    json.Number `json:"value"`
	Currency string      `json:"currency"`
}

func (v valueAmount) String() string { return money{Amount: v.Value, Currency: v.Currency}.String() }

type grossPrice struct {
	GrossValue json.Number `json:"grossValue"`
}

func (p grossPrice) in(currency string) string {
	return money{Amount: p.GrossValue, Currency: currency}.String()
}

type calculatedPrice struct {
	FinalPrice grossPrice `json:"finalPrice"`
}

func (p calculatedPrice) in(currency string) string { return p.FinalPrice.in(currency) }

type codeName struct {
	Code string    `json:"code"`
	Name localized `json:"name"`
}

func codeNameView(singular string) view[codeName] {
	return view[codeName]{
		singular:  singular,
		header:    []string{columnCode, columnName},
		localized: true,
		row:       func(c codeName, lang language) []string { return []string{c.Code, c.Name.in(lang)} },
	}
}

type validity struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type itemMetadata struct {
	Version    json.Number `json:"version"`
	CreatedAt  string      `json:"createdAt"`
	ModifiedAt string      `json:"modifiedAt"`
}

func spaced(parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, " ")
}

const maxValueRunes = 60

func cut(text string) string {
	if runes := []rune(text); len(runes) > maxValueRunes {
		return string(runes[:maxValueRunes-1]) + "…"
	}
	return text
}

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

func joinEach[T any](items []T, sep string, format func(T) string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = format(item)
	}
	return strings.Join(parts, sep)
}

func formatBool(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

func percent(value json.Number) string {
	if value == "" {
		return ""
	}
	return value.String() + "%"
}

func reported(value json.Number) string {
	if value == "" {
		return "not reported"
	}
	return value.String()
}

type personName struct {
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName"`
	LastName   string `json:"lastName"`
}

func (n personName) fullName() string { return spaced(n.FirstName, n.MiddleName, n.LastName) }
