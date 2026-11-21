package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

type sessionContext struct {
	SessionID      string `json:"sessionId"`
	CustomerID     string `json:"customerId"`
	SiteCode       string `json:"siteCode"`
	Currency       string `json:"currency"`
	Language       string `json:"language"`
	TargetLocation string `json:"targetLocation"`
	CartID         string `json:"cartId"`
}

const sessionContextManage = "sessioncontext.context_manage"

const sessionContexts = "/session-context/{tenant}/context"

const sessionContextPath = sessionContexts + "/" + placeholderID

func newSessionContextsCommand(a *app) *cobra.Command {
	const long = "Reads the context of a storefront session: its customer (once logged in, personal data; " +
		"ANONYMOUS before), site, currency, language, LOCATION (its target country) and active cart; " +
		"its custom attributes are only in `-o json`. Emporix creates the context with the session's " +
		"first cart; before that, the session's ID is not found (exit 4). The context of an " +
		"anonymous cart may name the tenant's default site rather than the cart's, and " +
		"names no cart once the cart was merged into a customer's."
	contexts := reader[sessionContext]{view: view[sessionContext]{
		singular: "session context",
		header:   []string{columnID, columnCustomer, columnSite, columnCurrency, "LANGUAGE", "LOCATION", "CART"},
		row: func(s sessionContext, _ language) []string {
			return []string{s.SessionID, s.CustomerID, s.SiteCode, s.Currency, s.Language, s.TargetLocation, s.CartID}
		},
	}, scopes: []string{sessionContextManage}}
	return newGroup("session-contexts", "Read and change session contexts", long, append([]*cobra.Command{
		contexts.command(a, &command{
			use:   "get " + columnID,
			short: "Show one session context",
			long:  long,
			req:   request{path: sessionContextPath},
			fetch: single{items: oneItem},
		}),
	}, sessionContextWrites(a)...)...)
}

func sessionContextWrites(a *app) []*cobra.Command {
	scopes := []string{sessionContextManage}
	const source = "SESSION_ID is the ID of `epx session-contexts get`; an unknown one is not found (exit 4). " +
		"Emporix does not check the version of the context for it."
	return []*cobra.Command{
		(&write{
			use: "update " + columnID, short: "Replace a session context", method: http.MethodPut,
			long: "Replaces the context of the storefront session ID with the JSON in --file, as " +
				"`epx session-contexts get ID -o json` prints it, with its metadata.version as read, which " +
				"Emporix requires. With --upsert, Emporix creates the context where the session has none " +
				"(an upsert). Emporix answers an anonymous or unknown session with 400 Bad Request (exit 1) " +
				"that requires the header saas-token, with or without --upsert, which epx does not send; a " +
				"session that a customer logged in to needs that customer's token too.",
			req: request{path: sessionContextPath}, scopes: scopes, body: &payload{},
			switches: []paramFlag{{param: "upsert", flag: "upsert", usage: "create the context where the session has none"}},
		}).command(a),
		newGroup("attributes", "Add and delete custom attributes of session contexts",
			"Adds and deletes the custom attributes of a session context, as its context member holds "+
				"them (`epx session-contexts get ID -o json`).",
			(&write{
				use: "create SESSION_ID", short: "Add a custom attribute to a session context", method: http.MethodPost,
				long: "Adds the custom attribute --key with the string --value to the context of the storefront " +
					"session SESSION_ID; --file gives the body instead, e.g. " +
					literal(`{"key": "k", "value": {"a": "b"}}`) + " for a " +
					"value that is a JSON object. " + source,
				req: request{path: sessionContextPath + "/attributes"}, scopes: scopes,
				body: &payload{typed: objectBody(requireAll([]bodyField{
					stringField("key", "key", "the `KEY` of the attribute"),
					stringField("value", "value", "the `VALUE` of the attribute, a string"),
				}))},
			}).command(a),
			(&write{
				use: "delete SESSION_ID KEY", short: "Delete a custom attribute of a session context",
				method: http.MethodDelete,
				long: "Deletes the custom attribute KEY of the context of the storefront session SESSION_ID. " +
					source,
				req:    request{path: sessionContexts + "/" + placeholderParent + "/attributes/" + placeholderID},
				scopes: scopes, body: noBody{},
			}).command(a),
		),
	}
}
