package cli

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// configurationManage is the scope that changes tenant and client
// configurations, the only one the reference names for their writes.
const configurationManage = "configuration.configuration_manage"

// newConfigurationWrites returns the writes of tenant configurations, whose
// KEY completes from keys, and those of the configurations of a client
// (settingOwner.writes), whose CLIENT completes from clients and KEY from
// clientKeys.
func newConfigurationWrites(a *app, keys, clients, clientKeys *idSource) []*cobra.Command {
	var commands []*cobra.Command
	for _, owner := range []*settingOwner{
		{
			some: "tenant configurations", one: "a tenant configuration",
			settings: "tenant configurations", setting: "the tenant configuration KEY",
			read: "get KEY", path: tenantConfigurationsPath, keys: keys,
		},
		{
			prefix: "client-", client: "CLIENT ",
			some: "configurations of a client", one: "a configuration of a client",
			settings: "configurations of the client CLIENT", setting: "the configuration KEY of the client CLIENT",
			read: "client-setting CLIENT KEY", path: clientConfigurationsPath, clients: clients, keys: clientKeys,
		},
	} {
		commands = append(commands, commandsOf(a, owner.writes())...)
	}
	return commands
}

// settingOwner is whose settings a set of configuration writes changes: the
// tenant's, or those of a client, which its writes then take as their first
// argument, CLIENT. Each owner has the same writes, named and described alike.
type settingOwner struct {
	prefix   string // of the names of its writes, e.g. "client-" in client-update
	client   string // its argument before KEY, followed by a space: "CLIENT ", or none
	some     string // its settings in a summary, e.g. "configurations of a client"
	one      string // one of its settings in a summary, e.g. "a configuration of a client"
	settings string // its settings in help, e.g. "configurations of the client CLIENT"
	setting  string // its setting KEY in help, e.g. "the configuration KEY of the client CLIENT"
	read     string // the read of its setting KEY, e.g. "client-setting CLIENT KEY"
	path     string // the collection of its settings

	// The sources of the IDs of its arguments, for completion: CLIENT's (none for the tenant's
	// settings, which take no CLIENT) and KEY's.
	clients, keys *idSource
}

// writes returns the writes of o's settings: create (POST of a JSON array of
// settings), update KEY (PUT of one whole setting) and delete KEY. A change of
// one value is an update of the whole setting, which a dry run shows without
// reading it first. The reference documents that a PUT of a key that was never set fails
// (400), but Emporix answers 404 (not found), so neither update nor set is an
// upsert: create makes a key.
func (o *settingOwner) writes() []*write {
	scopes := []string{configurationManage}
	item := request{path: o.path + "/" + placeholderID}
	epx := "`epx " + configurationsUse + " " // a command line in help, e.g. epx + o.read + "`"
	create := epx + o.prefix + "create`"
	return []*write{
		{
			use: strings.TrimSpace(o.prefix + "create " + o.client), short: "Create " + o.some,
			long: "Creates " + o.settings + " from the JSON array of settings that --file gives,\n" +
				"each with its key and value (and optionally secured, description, readOnly, restricted and\n" +
				"schemaUrl), and prints the settings created. Emporix answers 409 Conflict (exit 5) when one of\n" +
				"the keys exists already (the other keys of the request may be created then): change it with " +
				epx + o.prefix + "update`.",
			method: http.MethodPost, req: request{path: o.path}, scopes: scopes, body: &payload{},
			ids: argumentIDs{configurationClient: o.clients},
		},
		{
			use: o.prefix + "update " + o.client + "KEY", short: "Replace " + o.one,
			long: "Replaces " + o.setting + " with the setting that --file gives, a JSON\n" +
				"object with the key KEY (Emporix refuses another) and its value, and prints the setting. Give\n" +
				"every member to keep, such as secured, description and version, as\n" +
				epx + o.read + " -o json` shows them. Emporix refuses a key that was never set (404 Not Found,\n" +
				"exit 4): create it with " + create + ".",
			method: http.MethodPut, req: item, scopes: scopes, body: &payload{}, ids: o.ids(),
		},
		{
			use: o.prefix + "delete " + o.client + "KEY", short: "Delete " + o.one, method: http.MethodDelete,
			long: "Deletes " + o.setting + ", as\n" + epx + o.read + "` shows it; Emporix then applies its\n" +
				"default, if it has one. A key that is not set is not found (exit 4).",
			req: item, scopes: scopes, body: noBody{}, ids: o.ids(),
		},
	}
}

// ids returns the sources of the IDs of the arguments of a write of one of
// o's settings: CLIENT, where o takes it, and KEY.
func (o *settingOwner) ids() argumentIDs {
	return argumentIDs{configurationClient: o.clients, columnKey: o.keys}
}
