package cli

import "strings"

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

const locationColumns = "TYPE is HEADQUARTER, WAREHOUSE or OFFICE, ADDRESS the first address line and TAGS the " +
	"location's labels; phone numbers and the rest of the address are only in `-o json`."

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

const (
	locationMembers = "type (required: HEADQUARTER, WAREHOUSE or OFFICE), contactDetails (required: emails, " +
		"phones, addressLine1, addressLine2, city, state, postcode, countryCode and tags) and name; " +
		"leave out \"phones\": [] (an empty list), which Emporix refuses (exit 1)."
	locationRequired = "type and contactDetails"
)
