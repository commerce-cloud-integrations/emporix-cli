package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// webhookConfigurationWrites returns the writes of webhook configurations:
// create, update CODE (PUT, which creates no missing configuration), patch
// CODE (a JSON array of operations) and delete CODE (--force for an active
// one).
func webhookConfigurationWrites(a *app, configs resource[webhookConfiguration]) []*cobra.Command {
	scopes := []string{webhookManageScope}
	item := request{path: configs.itemPath()}
	code := "CODE is the code that `epx " + configs.listLine() + "` shows, such as http."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a webhook configuration", method: http.MethodPost,
			long: "Creates the webhook configuration in --file, a JSON object with its code, provider " +
				"(SVIX_SHARED, SVIX or HTTP), active and configuration: none for SVIX_SHARED, the apiKey of " +
				"your Svix account for SVIX, and for HTTP the destinationUrl, secretKey, headers and " +
				"eventsConfiguration (entries that send events of a type elsewhere, each without an id: " +
				"Emporix assigns one and refuses one given). For HTTP, Emporix checks that the destinationUrl " +
				"is reachable (400 Bad Request, exit 1, \"not reachable\" otherwise). Emporix answers with the " +
				"code. A tenant has at most one configuration per provider: when it has one, Emporix " +
				"answers 409 Conflict (exit 5); change that one with update or patch.",
			example: "  epx webhooks configurations create --file http.json --dry-run",
			req:     request{path: configs.path}, scopes: scopes, body: &payload{},
		}).command(a),
		(&write{
			use: configs.itemUse("update"), short: "Replace a webhook configuration", method: http.MethodPut,
			long: "Replaces the webhook configuration CODE with the one in --file, a JSON object as create " +
				"takes it. Keep the id of each entry of eventsConfiguration that stays, as " +
				"`epx webhooks configurations get CODE -o json` prints it; Emporix assigns an id to an entry " +
				"without one. A read gives secretKeyExists, not the secret: send secretKey (and apiKey for " +
				"SVIX) again, or change single members with patch. For HTTP, the destinationUrl must be reachable, " +
				"as at create. " + code,
			req: item, scopes: scopes, body: &payload{}, ids: configs.itemIDs(),
		}).command(a),
		(&write{
			use: configs.itemUse("patch"), short: "Change members of a webhook configuration", method: http.MethodPatch,
			long: "Changes the webhook configuration CODE with the JSON array of operations in --file, each " +
				"with op (UPSERT sets the member at path to value; REMOVE removes an item of a list), path " +
				"and value, e.g. [{\"op\":\"UPSERT\",\"path\":\"/active\",\"value\":false}]. For HTTP, an UPSERT of " +
				"/configuration/http/eventsConfigurationEntry adds an entry (without an id), and one of " +
				"/configuration/http/eventsConfigurationEntry/ID/active with false stops sending events to " +
				"the destination of the entry ID. " + code,
			example: "  echo '[{\"op\":\"UPSERT\",\"path\":\"/active\",\"value\":false}]' |\n" +
				"    epx webhooks configurations patch http --file - --yes",
			req: item, scopes: scopes, body: operationsFile(), ids: configs.itemIDs(),
		}).command(a),
		(&write{
			use: configs.itemUse("delete"), short: "Delete a webhook configuration", method: http.MethodDelete,
			long: "Deletes the webhook configuration CODE; Emporix deletes an active one only with --force. " +
				code,
			req: item, scopes: scopes, body: noBody{}, ids: configs.itemIDs(),
			switches: forceSwitch("delete the configuration also while it is active"),
		}).command(a),
	}
}

// subscriptionsBulkPatch subscribes the tenant to webhook events and
// unsubscribes it, from a JSON array of changes sent to path, the collection
// of its subscriptions, which Emporix answers per event with 207
// Multi-Status: a bulk write of up to bulkItemsDefault items, as the
// reference names no other limit, whose results name each item by its
// eventType, not by index.
func subscriptionsBulkPatch(path string) *write {
	return &write{
		use:   bulkPatchUse,
		short: "Subscribe to webhook events and unsubscribe from them",
		long: "Changes the tenant's webhook event subscriptions with the JSON array in --file, an item per " +
			"event: its eventType, action (SUBSCRIBE or UNSUBSCRIBE), the metadata.version of its " +
			"subscription, as `epx webhooks subscriptions list -o json` prints it (Emporix answers a " +
			"stale one with 409 Conflict), and optionally fieldsToUnsubscribe, the fields of the event " +
			"to leave out of its payload, and fieldsToSubscribe, those to send again. Each result " +
			"names the eventType of its item, as each failure that epx lists does.",
		example: "  epx webhooks subscriptions bulk-patch --file subscriptions.json --dry-run",
		method:  http.MethodPatch,
		req:     request{path: path},
		scopes:  []string{webhookManageScope},
		body:    &payload{maxItems: bulkItemsDefault, itemKey: "eventType"},
	}
}
