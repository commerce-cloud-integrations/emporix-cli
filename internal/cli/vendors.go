package cli

import (
	"slices"

	"github.com/spf13/cobra"
)

// vendorScope is the scope of every vendor read.
const vendorScope = "vendor.vendor_read"

// vendorManageScope is the scope of every write of vendors and their
// locations.
const vendorManageScope = "vendor.vendor_manage"

// vendor is the part of a Vendor Service vendor shown in tables.
type vendor struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	ParentID  string `json:"parentId"`
	LegalInfo struct {
		LegalName             string `json:"legalName"`
		CountryOfRegistration string `json:"countryOfRegistration"`
	} `json:"legalInfo"`
}

// vendors declares the vendors, whose IDs the VENDOR_ID of another read
// completes with too.
func vendors() resource[vendor] {
	return resource[vendor]{
		use:   "vendors",
		short: "Read and change vendors",
		about: "Reads and changes vendors, the sellers of a marketplace tenant; `epx vendor-locations`\n" +
			"reads and changes their locations.\n" + versionedWritesNote,
		long: "Reads vendors, the sellers of a marketplace tenant. TYPE is COMPANY or SUBSIDIARY, PARENT\n" +
			"the vendor a subsidiary belongs to and COUNTRY the country of registration. Registration\n" +
			"details, contact details and locations are only in -o json; `epx vendor-locations` lists\n" +
			"the locations.",
		path:    "/vendor/{tenant}/vendors",
		scopes:  []string{vendorScope},
		filters: listFilters{kinds: filterable},
		view: view[vendor]{
			singular: "vendor",
			header:   []string{columnID, columnName, columnType, columnParent, "LEGAL NAME", columnCountry},
			row: func(v vendor, _ language) []string {
				return []string{v.ID, v.Name, v.Type, v.ParentID, v.LegalInfo.LegalName, v.LegalInfo.CountryOfRegistration}
			},
		},
	}
}

// newVendorsCommand reads vendors, the sellers of a marketplace tenant, also
// those a search finds, and creates, upserts and deletes them.
func newVendorsCommand(a *app) *cobra.Command {
	r := vendors()
	reads := []*cobra.Command{r.searchCommand(a,
		"Reads the vendors that an Emporix q filter finds (--query; without it, every vendor).",
		queryField())}
	writes := r.versionedWrites(a, []string{vendorManageScope}, &versioned{
		members: "name (required), type (COMPANY, the default, or SUBSIDIARY), parentId (the vendor a\n" +
			"subsidiary belongs to), legalInfo, contactDetails (a list of contacts, each as\n" +
			"{\"email\": ...}, which is all Emporix keeps of one), locations (its vendor locations, each\n" +
			"as {\"id\": ...} or in full, as get prints it) and mixins; parentId and locations must name\n" +
			"items that exist (400 Bad Request, exit 1, otherwise).",
		required:   "name and type",
		deleteNote: "Its subsidiaries stay and still name it as their parentId.",
		unknown:    answeredDeleted,
	})
	return newResourceCommand(a, r, slices.Concat(reads, writes)...)
}

// newVendorLocationsCommand reads the locations of vendors, and creates,
// upserts and deletes them.
func newVendorLocationsCommand(a *app) *cobra.Command {
	locations := resource[location]{
		use:   "vendor-locations",
		short: "Read and change the locations of vendors",
		about: "Reads and changes the locations of vendors, contact emails included. A vendor names its\n" +
			"locations in locations, each as {\"id\": ...}.\n" + versionedWritesNote,
		long:    "Reads the locations of vendors, contact emails included." + locationColumns,
		path:    "/vendor/{tenant}/locations",
		scopes:  []string{vendorScope},
		filters: listFilters{kinds: filterable},
		view:    locationView("vendor location"),
	}
	writes := locations.versionedWrites(a, []string{vendorManageScope}, &versioned{
		members: locationMembers, required: locationRequired,
		deleteNote: "A vendor that names it in locations no longer does.",
	})
	return newResourceCommand(a, locations, writes...)
}
