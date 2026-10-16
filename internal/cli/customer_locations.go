package cli

import "github.com/spf13/cobra"

// newCustomerLocationsCommand reads the locations of B2B customers, such as
// their headquarters and warehouses, and creates, upserts and deletes them.
func newCustomerLocationsCommand(a *app) *cobra.Command {
	locations := resource[location]{
		use:   "customer-locations",
		short: "Read and change the locations of B2B customers",
		about: "Reads and changes the locations of B2B customers (legal entities), contact emails\n" +
			"included. A legal entity names its locations in entitiesAddresses, each as {\"id\": ...}.\n" +
			versionedWritesNote,
		long:    "Reads the locations of B2B customers (legal entities), contact emails included." + locationColumns,
		path:    "/customer-management/{tenant}/locations",
		scopes:  []string{"customermanagement.location_read"},
		filters: listFilters{kinds: filterable},
		view:    locationView("customer location"),
	}
	writes := locations.versionedWrites(a, []string{"customermanagement.location_manage"}, &versioned{
		members: locationMembers, required: locationRequired,
		deleteNote: "A legal entity that names it in entitiesAddresses no longer does.",
	})
	return newResourceCommand(a, locations, writes...)
}
