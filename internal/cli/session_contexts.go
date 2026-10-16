package cli

import (
	"net/http"

	"github.com/spf13/cobra"
)

// sessionContext is the part of a Session Context Service session context
// shown in tables.
type sessionContext struct {
	SessionID      string `json:"sessionId"`
	CustomerID     string `json:"customerId"`
	SiteCode       string `json:"siteCode"`
	Currency       string `json:"currency"`
	Language       string `json:"language"`
	TargetLocation string `json:"targetLocation"`
	CartID         string `json:"cartId"`
}

// sessionContextManage is the one scope of the Session Context Service that
// a service token may hold: it reads and changes session contexts.
const sessionContextManage = "sessioncontext.context_manage"

// sessionContexts is the path of the session contexts, below which each
// storefront session has its own.
const sessionContexts = "/session-context/{tenant}/context"

// sessionContextPath is the session context of the storefront session that
// {id} names.
const sessionContextPath = sessionContexts + "/" + placeholderID

// newSessionContextsCommand reads the session context of a storefront
// session, which Emporix reads only with the write-capable
// sessioncontext.context_manage scope, so the read is opt-in, and changes it
// (sessionContextWrites). Emporix lists no session contexts, so the group
// reads with get only.
func newSessionContextsCommand(a *app) *cobra.Command {
	const long = "Reads the context of a storefront session: its customer (once logged in, personal data;\n" +
		"ANONYMOUS before), site, currency, language, LOCATION (its target country) and active cart;\n" +
		"its custom attributes are only in -o json. Emporix creates the context with the session's\n" +
		"first cart; before that, the session's ID is not found (exit 4). The context of an\n" +
		"anonymous cart may name the tenant's default site rather than the cart's, and\n" +
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

// sessionContextWrites returns the writes of a session context: update, and
// the group of its custom attributes, which adds and deletes one.
func sessionContextWrites(a *app) []*cobra.Command {
	scopes := []string{sessionContextManage}
	const source = "\nSESSION_ID is the ID of `epx session-contexts get`; an unknown one is not found (exit 4).\n" +
		"Emporix does not check the version of the context for it."
	return []*cobra.Command{
		(&write{
			use: "update " + columnID, short: "Replace a session context", method: http.MethodPut,
			long: "Replaces the context of the storefront session ID with the JSON in --file, as\n" +
				"`epx session-contexts get ID -o json` prints it, with its metadata.version as read, which\n" +
				"Emporix requires. With --upsert, Emporix creates the context where the session has none\n" +
				"(an upsert). Emporix answers an anonymous or unknown session with 400 Bad Request (exit 1)\n" +
				"that requires the header saas-token, with or without --upsert, which epx does not send; a\n" +
				"session that a customer logged in to needs that customer's token too.",
			req: request{path: sessionContextPath}, scopes: scopes, body: &payload{},
			switches: []paramFlag{{param: "upsert", flag: "upsert", usage: "create the context where the session has none"}},
		}).command(a),
		newGroup("attributes", "Add and delete custom attributes of session contexts",
			"Adds and deletes the custom attributes of a session context, as its context member holds\n"+
				"them (`epx session-contexts get ID -o json`).",
			(&write{
				use: "create SESSION_ID", short: "Add a custom attribute to a session context", method: http.MethodPost,
				long: "Adds the custom attribute --key with the string --value to the context of the storefront\n" +
					"session SESSION_ID; --file gives the body instead, e.g. {\"key\": \"k\", \"value\": {\"a\": \"b\"}} for a\n" +
					"value that is a JSON object." + source,
				req: request{path: sessionContextPath + "/attributes"}, scopes: scopes,
				body: &payload{typed: objectBody(requireAll([]bodyField{
					stringField("key", "key", "the `KEY` of the attribute"),
					stringField("value", "value", "the `VALUE` of the attribute, a string"),
				}))},
			}).command(a),
			(&write{
				use: "delete SESSION_ID KEY", short: "Delete a custom attribute of a session context",
				method: http.MethodDelete,
				long: "Deletes the custom attribute KEY of the context of the storefront session SESSION_ID." +
					source,
				req:    request{path: sessionContexts + "/" + placeholderParent + "/attributes/" + placeholderID},
				scopes: scopes, body: noBody{},
			}).command(a),
		),
	}
}
