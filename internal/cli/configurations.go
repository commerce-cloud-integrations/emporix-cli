package cli

import (
	"encoding/json"
	"slices"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// configuration is the part of a Configuration Service tenant setting shown in
// tables.
type configuration struct {
	Key      string          `json:"key"`
	Value    json.RawMessage `json:"value"`
	Version  json.Number     `json:"version"`
	Secured  *bool           `json:"secured"`
	ReadOnly *bool           `json:"readOnly"`
}

// configurationView shows a setting by its key, value (long: tables cut it)
// and flags.
func configurationView(singular string) view[configuration] {
	return view[configuration]{
		singular: singular,
		header:   []string{columnKey, columnValue, columnVersion, "SECURED", "READ ONLY"},
		long:     []string{columnValue},
		row: func(c configuration, _ language) []string {
			return []string{c.Key, valueCell(c.Value), c.Version.String(), formatBool(c.Secured), formatBool(c.ReadOnly)}
		},
	}
}

// configurationScope reads tenant and client configurations.
const configurationScope = "configuration.configuration_view"

// The collections of tenant configurations, of the clients that have
// configurations, and of the configurations of the client that a command's
// first argument, CLIENT, names.
const (
	tenantConfigurationsPath = "/configuration/{tenant}/configurations"
	configurationClientsPath = "/configuration/{tenant}/clients"
	clientConfigurationsPath = configurationClientsPath + "/" + placeholderParent + "/configurations"
)

// configurationClient is the column of the clients that have configurations,
// and the argument that names one.
const configurationClient = "CLIENT"

// newConfigurationsCommand reads tenant configurations, the global ones and
// those of clients, and writes those of the tenant and of clients
// (newConfigurationWrites). The service returns each list whole (live, it
// ignores pageNumber and pageSize), so each is read with one request.
func newConfigurationsCommand(a *app) *cobra.Command {
	const values = "Tables show values as sent, secured ones included, and cut long ones (-o json prints\n" +
		"them whole)."
	whole := listing{read: (*emporix.Client).GetUnpaged}
	global := reader[configuration]{view: configurationView("global configuration")}
	clients := reader[string]{
		view:   stringsView("configuration client", configurationClient),
		scopes: []string{configurationScope},
	}
	clientSettings := reader[configuration]{
		view:   configurationView("client configuration"),
		scopes: []string{configurationScope},
	}
	clientList, settingList := request{path: configurationClientsPath}, request{path: clientConfigurationsPath}
	clientIDs, keyIDs := listedIDs(clients, clientList, whole, configurationClient),
		listedIDs(clientSettings, settingList, whole, columnKey)
	reads := []*cobra.Command{
		global.command(a, &command{
			use:   "global",
			short: "List global configurations",
			long:  "Reads global configurations, which hold for every tenant.\n" + values,
			req:   request{path: "/configuration/{tenant}/global-configurations"},
			fetch: whole,
		}),
		clients.command(a, &command{
			use:   "clients",
			short: "List the clients that have configurations",
			long:  "Reads the clients (such as a storefront or an extension) that keep their own settings.",
			req:   clientList,
			fetch: whole,
		}),
		clientSettings.command(a, &command{
			use:   "for-client CLIENT",
			short: "List the configurations of a client",
			long: "Reads the settings of a client.\n" + values + "\n" +
				"CLIENT is a name that `epx configurations clients` shows; an unknown one lists nothing.",
			req:   settingList,
			fetch: whole,
			ids:   argumentIDs{configurationClient: clientIDs},
		}),
		clientSettings.command(a, &command{
			use:   "client-setting CLIENT KEY",
			short: "Show one configuration of a client",
			long: "Reads one setting of a client.\n" + values + "\n" +
				"CLIENT is a name that `epx configurations clients` shows and KEY a key that\n" +
				"`epx configurations for-client CLIENT` shows; an unknown one is not found (exit 4).",
			req:   request{path: clientConfigurationsPath + "/" + placeholderID},
			fetch: single{items: oneItem},
			ids:   argumentIDs{configurationClient: clientIDs, columnKey: keyIDs},
		}),
	}
	tenant := resource[configuration]{
		use:   configurationsUse,
		short: "Read and change tenant configurations",
		long: "Reads tenant configurations (settings such as project_lang).\n" + values + "\n" +
			"A setting that was never set is not found (exit 4): Emporix then applies its default.\n" +
			"`epx configurations global`, `clients`, `for-client` and `client-setting` read global\n" +
			"settings and those of clients. `epx configurations create`, `update` and `delete`\n" +
			"change settings, and `client-create`, `client-update` and `client-delete` those\n" +
			"of a client.",
		path:   tenantConfigurationsPath,
		scopes: []string{configurationScope},
		read:   (*emporix.Client).GetUnpaged,
		view:   configurationView("configuration"),
	}
	writes := newConfigurationWrites(a, tenant.ids(), clientIDs, keyIDs)
	return newResourceCommand(a, tenant, slices.Concat(reads, writes)...)
}
