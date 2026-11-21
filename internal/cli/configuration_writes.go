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
			summaryPlural: "tenant configurations", summaryNoun: "a tenant configuration",
			helpPlural: "tenant configurations", helpNoun: "the tenant configuration KEY",
			read: "get KEY", path: tenantConfigurationsPath, keys: keys,
		},
		{
			prefix: "client-", client: "CLIENT ",
			summaryPlural: "configurations of a client", summaryNoun: "a configuration of a client",
			helpPlural: "configurations of the client CLIENT", helpNoun: "the configuration KEY of the client CLIENT",
			read: "client-setting CLIENT KEY", path: clientConfigurationsPath, clients: clients, keys: clientKeys,
		},
	} {
		commands = append(commands, commandsOf(a, owner.writes())...)
	}
	return commands
}

type settingOwner struct {
	prefix        string
	client        string
	summaryPlural string
	summaryNoun   string
	helpPlural    string
	helpNoun      string
	read          string
	path          string

	clients, keys *idSource
}

func (o *settingOwner) writes() []*write {
	scopes := []string{configurationManage}
	item := request{path: o.path + "/" + placeholderID}
	epx := "`epx " + configurationsUse + " "
	create := epx + o.prefix + "create`"
	return []*write{
		{
			use: strings.TrimSpace(o.prefix + "create " + o.client), short: "Create " + o.summaryPlural,
			long: "Creates " + o.helpPlural + " from the JSON array of settings that --file gives, " +
				"each with its key and value (and optionally secured, description, readOnly, restricted and " +
				"schemaUrl), and prints the settings created. Emporix answers 409 Conflict (exit 5) when one of " +
				"the keys exists already (the other keys of the request may be created then): change it with " +
				epx + o.prefix + "update`.",
			method: http.MethodPost, req: request{path: o.path}, scopes: scopes, body: &payload{},
			ids: argumentIDs{configurationClient: o.clients},
		},
		{
			use: o.prefix + "update " + o.client + "KEY", short: "Replace " + o.summaryNoun,
			long: "Replaces " + o.helpNoun + " with the setting that --file gives, a JSON " +
				"object with the key KEY (Emporix refuses another) and its value, and prints the setting. Give " +
				"every member to keep, such as secured, description and version, as " +
				epx + o.read + " -o json` shows them. Emporix refuses a key that was never set (404 Not Found, " +
				"exit 4): create it with " + create + ".",
			method: http.MethodPut, req: item, scopes: scopes, body: &payload{}, ids: o.ids(),
		},
		{
			use: o.prefix + "delete " + o.client + "KEY", short: "Delete " + o.summaryNoun, method: http.MethodDelete,
			long: "Deletes " + o.helpNoun + ", as " + epx + o.read + "` shows it; Emporix then applies its " +
				"default, if it has one. A key that is not set is not found (exit 4).",
			req: item, scopes: scopes, body: noBody{}, ids: o.ids(),
		},
	}
}

func (o *settingOwner) ids() argumentIDs {
	return argumentIDs{configurationClient: o.clients, columnKey: o.keys}
}
