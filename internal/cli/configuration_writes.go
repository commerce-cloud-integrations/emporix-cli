package cli

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

const configurationManage = "configuration.configuration_manage"

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

type settingOwner struct {
	prefix   string
	client   string
	some     string
	one      string
	settings string
	setting  string
	read     string
	path     string

	clients, keys *idSource
}

func (o *settingOwner) writes() []*write {
	scopes := []string{configurationManage}
	item := request{path: o.path + "/" + placeholderID}
	epx := "`epx " + configurationsUse + " "
	create := epx + o.prefix + "create`"
	return []*write{
		{
			use: strings.TrimSpace(o.prefix + "create " + o.client), short: "Create " + o.some,
			long: "Creates " + o.settings + " from the JSON array of settings that --file gives, " +
				"each with its key and value (and optionally secured, description, readOnly, restricted and " +
				"schemaUrl), and prints the settings created. Emporix answers 409 Conflict (exit 5) when one of " +
				"the keys exists already (the other keys of the request may be created then): change it with " +
				epx + o.prefix + "update`.",
			method: http.MethodPost, req: request{path: o.path}, scopes: scopes, body: &payload{},
			ids: argumentIDs{configurationClient: o.clients},
		},
		{
			use: o.prefix + "update " + o.client + "KEY", short: "Replace " + o.one,
			long: "Replaces " + o.setting + " with the setting that --file gives, a JSON " +
				"object with the key KEY (Emporix refuses another) and its value, and prints the setting. Give " +
				"every member to keep, such as secured, description and version, as " +
				epx + o.read + " -o json` shows them. Emporix refuses a key that was never set (404 Not Found, " +
				"exit 4): create it with " + create + ".",
			method: http.MethodPut, req: item, scopes: scopes, body: &payload{}, ids: o.ids(),
		},
		{
			use: o.prefix + "delete " + o.client + "KEY", short: "Delete " + o.one, method: http.MethodDelete,
			long: "Deletes " + o.setting + ", as " + epx + o.read + "` shows it; Emporix then applies its " +
				"default, if it has one. A key that is not set is not found (exit 4).",
			req: item, scopes: scopes, body: noBody{}, ids: o.ids(),
		},
	}
}

func (o *settingOwner) ids() argumentIDs {
	return argumentIDs{configurationClient: o.clients, columnKey: o.keys}
}
