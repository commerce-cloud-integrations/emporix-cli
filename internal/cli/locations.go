package cli

import "strings"

// location is the part of a location shown in tables; customers (legal
// entities) and vendors have locations of the same shape.
type location struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	ContactDetails struct {
		Emails       []string `json:"emails"`
		AddressLine1 string   `json:"addressLine1"`
		Postcode     string   `json:"postcode"`
		City         string   `json:"city"`
		CountryCode  string   `json:"countryCode"`
		Tags         []string `json:"tags"`
	} `json:"contactDetails"`
}

// locationColumns ends the help of every location read.
const locationColumns = "\nTYPE is HEADQUARTER, WAREHOUSE or OFFICE, ADDRESS the first address line and TAGS the\n" +
	"location's labels; phone numbers and the rest of the address are only in -o json."

// locationView is the table of every location read; singular names its locations.
func locationView(singular string) view[location] {
	return view[location]{
		singular: singular,
		header:   []string{columnID, columnName, columnType, "ADDRESS", "POSTCODE", "CITY", columnCountry, "EMAILS", "TAGS"},
		row: func(l location, _ language) []string {
			d := l.ContactDetails
			return []string{
				l.ID, l.Name, l.Type, d.AddressLine1, d.Postcode, d.City, d.CountryCode, strings.Join(d.Emails, ","),
				strings.Join(d.Tags, ","),
			}
		},
	}
}

// locationMembers are the members of the body of a location's create, which
// customers (legal entities) and vendors make alike; locationRequired those
// that its update requires.
const (
	locationMembers = "type (required: HEADQUARTER, WAREHOUSE or OFFICE), contactDetails (required: emails,\n" +
		"phones, addressLine1, addressLine2, city, state, postcode, countryCode and tags) and name;\n" +
		"leave out \"phones\": [] (an empty list), which Emporix refuses (exit 1)."
	locationRequired = "type and contactDetails"
)
