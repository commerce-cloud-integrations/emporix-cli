package cli

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// webhookConfiguration is the part of a webhook configuration shown in tables.
type webhookConfiguration struct {
	Code          string `json:"code"`
	Provider      string `json:"provider"`
	Active        *bool  `json:"active"`
	Configuration struct {
		DestinationURL      string            `json:"destinationUrl"`
		EventsConfiguration []json.RawMessage `json:"eventsConfiguration"` // counted only
	} `json:"configuration"`
}

// eventSubscription is a webhook event type and whether the tenant receives it.
type eventSubscription struct {
	Event struct {
		Type string    `json:"type"`
		Name localized `json:"name"`
	} `json:"event"`
	Subscription string `json:"subscription"`
}

// webhookStatistics is how many webhooks a tenant may send and sent per month.
type webhookStatistics struct {
	WebhooksLimit   json.Number       `json:"webhooksLimit"`
	MonthStatistics []monthStatistics `json:"monthStatistics"`
}

// monthStatistics is how many webhooks were sent in a month.
type monthStatistics struct {
	Month string      `json:"statisticsMonth"`
	Sent  json.Number `json:"succesfullySentWebhooks"` // spelled so by Emporix
}

func (m monthStatistics) String() string { return m.Month + ": " + m.Sent.String() }

// dashboardAccess is a login link to a tenant's webhook dashboard.
type dashboardAccess struct {
	URL string `json:"url"`
}

// The scopes of the Webhook Service: webhookScope reads webhook
// configurations, subscriptions and statistics; webhookManageScope, the only
// one the reference names for the writes, changes configurations and
// subscriptions, and reads the dashboard access, which Emporix gives only
// with it.
const (
	webhookScope       = "webhook.subscription_read"
	webhookManageScope = "webhook.subscription_manage"
)

// newWebhooksCommand reads the tenant's webhook setup: where events go, which
// events it subscribes to, how many webhooks it sent, and a link to its
// webhook dashboard, which Emporix gives only with a write-capable scope
// (opt-in).
func newWebhooksCommand(a *app) *cobra.Command {
	statistics := reader[webhookStatistics]{view: view[webhookStatistics]{
		singular: "webhook statistics",
		header:   []string{"LIMIT", "SENT PER MONTH"},
		row: func(s webhookStatistics, _ language) []string {
			return []string{reported(s.WebhooksLimit), joinEach(s.MonthStatistics, ", ",
				monthStatistics.String)}
		},
	}, scopes: []string{webhookScope}}
	dashboard := reader[dashboardAccess]{view: view[dashboardAccess]{
		singular: "dashboard access",
		header:   []string{columnURL},
		row:      func(d dashboardAccess, _ language) []string { return []string{d.URL} },
	}, scopes: []string{webhookManageScope}}
	return newGroup("webhooks", "Read and change webhook configurations and subscriptions; read statistics",
		"Reads the tenant's webhooks: where events are sent, which events it subscribes to, how many\n"+
			"webhooks it sent, and a link to its webhook dashboard; changes where events are sent and\n"+
			"which events it subscribes to.",
		newWebhookConfigurationsCommand(a),
		newWebhookSubscriptionsCommand(a),
		statistics.command(a, &command{
			use:   "statistics",
			short: "Show how many webhooks were sent per month",
			long: "Reads the monthly webhook LIMIT and the webhooks sent per month, for the months Emporix\n" +
				"reports by default. For other months, run\n" +
				"`epx api get /webhook/{tenant}/statistics -p fromYearMonth=2026-01 -p toYearMonth=2026-06 --scope " +
				webhookScope + "`.",
			req:   request{path: "/webhook/{tenant}/statistics"},
			fetch: single{items: oneObject},
		}),
		dashboard.command(a, &command{
			use:   "dashboard-access",
			short: "Show a link that logs in to the webhook dashboard",
			long: "Reads a URL that logs in to the tenant's webhook dashboard, the Svix application portal of\n" +
				"a tenant whose webhooks Emporix sends through its shared Svix account (provider\n" +
				"SVIX_SHARED). The URL holds its own authentication, so treat it as a secret; -o json also\n" +
				"prints the portal's token.",
			req:   request{path: "/webhook/{tenant}/dashboard-access"},
			fetch: single{items: oneItem},
		}),
	)
}

// newWebhookSubscriptionsCommand reads the webhook events and whether the
// tenant subscribes to them, and changes its subscriptions.
func newWebhookSubscriptionsCommand(a *app) *cobra.Command {
	const path = "/webhook/{tenant}/event-subscriptions"
	subscriptions := reader[eventSubscription]{view: view[eventSubscription]{
		singular:  "event subscription",
		header:    []string{"EVENT", columnStatus, columnName},
		localized: true,
		row: func(s eventSubscription, lang language) []string {
			return []string{s.Event.Type, s.Subscription, s.Event.Name.in(lang)}
		},
	}, scopes: []string{webhookScope}}
	return newGroup("subscriptions", "Read and change the webhook events that the tenant subscribes to",
		"Reads the webhook events and whether the tenant subscribes to them, and subscribes to\n"+
			"events or unsubscribes from them.",
		subscriptions.command(a, &command{
			use:   listUse,
			short: "List the webhook events and whether the tenant subscribes to them",
			long: "Reads every webhook event type and whether the tenant receives it: STATUS is SUBSCRIBED,\n" +
				"UNSUBSCRIBED or NONE. The service returns them whole, so they are read with one request.",
			req:   request{path: path},
			fetch: listing{read: (*emporix.Client).GetUnpaged},
		}),
		subscriptionsBulkPatch(path).command(a))
}

// newWebhookConfigurationsCommand reads and changes webhook configurations.
func newWebhookConfigurationsCommand(a *app) *cobra.Command {
	configs := resource[webhookConfiguration]{
		use:   configurationsUse,
		noun:  "webhook configurations",
		short: "Read, create, change and delete webhook configurations",
		under: "webhooks",
		long: "Reads webhook configurations: PROVIDER is SVIX_SHARED, SVIX or HTTP, DESTINATION where\n" +
			"events are sent and ENTRIES counts the event entries, each with its own destination,\n" +
			"headers and filter (inactive ones included), which only -o json shows. The service returns\n" +
			"the configurations whole, so they are read with one request.",
		path:   "/webhook/{tenant}/config",
		scopes: []string{webhookScope},
		read:   (*emporix.Client).GetUnpaged,
		view: view[webhookConfiguration]{
			singular: "webhook configuration",
			header:   []string{columnCode, columnProvider, columnActive, "DESTINATION", "ENTRIES"},
			row: func(c webhookConfiguration, _ language) []string {
				return []string{
					c.Code, c.Provider, formatBool(c.Active), c.Configuration.DestinationURL,
					strconv.Itoa(len(c.Configuration.EventsConfiguration)),
				}
			},
		},
	}
	return newResourceCommand(a, configs, webhookConfigurationWrites(a, configs)...)
}
