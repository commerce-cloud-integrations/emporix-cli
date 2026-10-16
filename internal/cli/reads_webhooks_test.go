package cli

// webhookReads returns the Webhook Service reads.
func webhookReads() []readCase {
	return []readCase{
		{
			// The configurations come whole; DESTINATION is the URL as sent; ENTRIES counts the event
			// entries.
			args:   []string{"webhooks", "configurations", "list"},
			path:   "/webhook/acme1/config",
			scopes: []string{"webhook.subscription_read"},
			mode:   getWhole,
			body: `[{"code":"svix","provider":"SVIX_SHARED","active":true},{"code":"erp","provider":"HTTP",
				"active":false,"configuration":{"destinationUrl":"https://erp.example.com/hook?code=K2#f","secretKeyExists":true,
				"eventsConfiguration":[{"id":"e1","eventType":"order.created"}]}}]`,
			want: "CODE  PROVIDER     ACTIVE  DESTINATION                             ENTRIES\n" +
				"svix  SVIX_SHARED  true                                            0\n" +
				"erp   HTTP         false   https://erp.example.com/hook?code=K2#f  1\n",
		},
		{
			args:   []string{"webhooks", "configurations", "get", "erp"},
			path:   "/webhook/acme1/config/erp",
			scopes: []string{"webhook.subscription_read"},
			mode:   getWhole,
			body:   `{"code":"erp","provider":"HTTP","active":true,"configuration":{"destinationUrl":"https://x.example"}}`,
			want: "CODE  PROVIDER  ACTIVE  DESTINATION        ENTRIES\n" +
				"erp   HTTP      true    https://x.example  0\n",
		},
		{
			args:   []string{"webhooks", "subscriptions", "list", "--language", "de"},
			path:   "/webhook/acme1/event-subscriptions",
			scopes: []string{"webhook.subscription_read"},
			mode:   getWhole,
			body: `[{"event":{"type":"order.created","name":{"en":"Order created","de":"Bestellung erstellt"}},
				"subscription":"SUBSCRIBED"},{"event":{"type":"cart.closed","name":{"en":"Cart closed"}},"subscription":"NONE"}]`,
			want: "EVENT          STATUS      NAME\n" +
				"order.created  SUBSCRIBED  Bestellung erstellt\n" +
				"cart.closed    NONE        Cart closed (en)\n",
		},
		{
			// One row: the limit and the webhooks sent per month.
			args:   []string{"webhooks", "statistics"},
			path:   "/webhook/acme1/statistics",
			scopes: []string{"webhook.subscription_read"},
			mode:   getWhole,
			body: `{"webhooksLimit":5000,"monthStatistics":[{"statisticsMonth":"2026-09","succesfullySentWebhooks":120},
				{"statisticsMonth":"2026-10","succesfullySentWebhooks":80}]}`,
			want: "LIMIT  SENT PER MONTH\n5000   2026-09: 120, 2026-10: 80\n",
		},
		{
			// Opt-in; the token is only in -o json.
			args:   []string{"webhooks", "dashboard-access", "--allow-write-scope"},
			path:   "/webhook/acme1/dashboard-access",
			scopes: []string{"webhook.subscription_manage"},
			mode:   getWhole,
			body:   `{"url":"https://app.svix.example/login#key=K1","token":"appsk_T1"}`,
			want:   "URL\nhttps://app.svix.example/login#key=K1\n",
		},
	}
}
