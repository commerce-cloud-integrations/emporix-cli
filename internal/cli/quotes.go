package cli

import (
	"cmp"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

const quoteScope = "quote.quote_read"

type quote struct {
	ID     string `json:"id"`
	Status struct {
		Value string `json:"value"`
	} `json:"status"`
	Customer struct {
		personName
		CustomerID   string `json:"customerId"`
		ContactEmail string `json:"contactEmail"`
	} `json:"customer"`
	Company struct {
		Name string `json:"name"`
	} `json:"company"`
	SiteCode   string `json:"siteCode"`
	ValidTo    string `json:"validTo"`
	Currency   string `json:"currency"`
	TotalPrice struct {
		GrossValue json.Number `json:"grossValue"`
		Currency   string      `json:"currency"`
	} `json:"totalPrice"`
}

func (q *quote) customer() string {
	return cmp.Or(q.Customer.ContactEmail, q.Customer.fullName(), q.Customer.CustomerID)
}

func (q *quote) total() string {
	return money{Amount: q.TotalPrice.GrossValue, Currency: cmp.Or(q.TotalPrice.Currency, q.Currency)}.String()
}

type quoteChange struct {
	ID            string `json:"id"`
	ModifiedAt    string `json:"modifiedAt"`
	UserType      string `json:"userType"`
	UserFirstName string `json:"userFirstName"`
	UserLastName  string `json:"userLastName"`
	Op            string `json:"op"`
	Path          string `json:"path"`
}

type quoteReason struct {
	ID      string    `json:"id"`
	Code    string    `json:"code"`
	Type    string    `json:"type"`
	Message localized `json:"message"`
}

func newQuotesCommand(a *app) *cobra.Command {
	quotes := resource[quote]{
		use:   "quotes",
		short: "Read, create and change quotes, their PDF, history and reasons",
		long: "Reads quotes, personal data included. STATUS is CREATING, AWAITING, OPEN, IN_PROGRESS, " +
			"ACCEPTED, DECLINED, DECLINED_BY_MERCHANT or EXPIRED, CUSTOMER the customer it is for (email, " +
			"name or customer id), and TOTAL the gross total in the quote's currency. Items, prices, " +
			"addresses and comments are only in `-o json`. " +
			"`epx quotes reasons` lists the reasons for changing or declining a quote. " +
			"`epx quotes create`, `patch` and `delete` change quotes; `pdf` saves a quote's PDF.",
		path:    "/quote/{tenant}/quotes",
		scopes:  []string{quoteScope},
		filters: listFilters{kinds: filterable},
		view: view[quote]{
			singular: "quote",
			header:   []string{columnID, columnStatus, columnCustomer, columnCompany, columnSite, "VALID TO", columnTotal},
			row: func(q quote, _ language) []string {
				return []string{q.ID, q.Status.Value, q.customer(), q.Company.Name, q.SiteCode, q.ValidTo, q.total()}
			},
		},
	}
	history := childCommand(a, quotes, child[quoteChange]{
		use:   "history",
		short: "List the changes of a quote",
		long: "Reads the history of a quote, personal data included: when (MODIFIED), by whom (BY: user " +
			"type and name) and what changed (CHANGE: the operation and the changed part). Old and new " +
			"values are only in `-o json`. Emporix sends the whole history in one response.",
		fetch: listing{read: (*emporix.Client).GetUnpaged},
		view: view[quoteChange]{
			singular: "quote change",
			header:   []string{columnID, columnModified, "BY", "CHANGE"},
			row: func(c quoteChange, _ language) []string {
				return []string{c.ID, c.ModifiedAt, spaced(c.UserType, c.UserFirstName, c.UserLastName), spaced(c.Op, c.Path)}
			},
		},
	})
	return newResourceCommand(a, quotes, append([]*cobra.Command{history, newQuotePDFCommand(a, quotes),
		newQuoteReasonsCommand(a)}, commandsOf(a, quoteWrites(quotes))...)...)
}

func newQuotePDFCommand(a *app, quotes resource[quote]) *cobra.Command {
	return quotes.reader().fileCommand(a, &fileRead{command: command{
		use: "pdf " + quotes.argument(), short: "Save the PDF of a quote",
		long: "Reads the PDF of the quote " + quotes.argument() + "; an unknown quote is not found (exit 4). " +
			quotes.argumentSource() + ". In a --language for which a product has no name, the PDF shows the template " +
			"variable $quoteitem.productName in its place; a language that the tenant does not configure is " +
			"answered in English, not refused.",
		req: request{path: quotes.itemPath() + "/pdf"}, ids: argumentIDs{quotes.argument(): quotes.ids()},
	}, example: "  epx quotes pdf Q1001 --out Q1001.pdf --language de\n  epx quotes pdf Q1001 --out - | lpr",
		noun: "PDF", media: "application/pdf"})
}

func newQuoteReasonsCommand(a *app) *cobra.Command {
	reasons := resource[quoteReason]{
		use:   "reasons",
		noun:  "quote reasons",
		under: "quotes",
		short: "Read, create, change and delete quote reasons",
		long: "Reads the reasons a customer or employee can give when asking for changes to a quote " +
			"(TYPE CHANGE) or declining it (DECLINE), and their MESSAGE. " +
			"`epx quotes reasons create`, `update` and `delete` change them.",
		path:    "/quote/{tenant}/quote-reasons",
		scopes:  []string{quoteScope},
		filters: listFilters{kinds: filterable},
		view: view[quoteReason]{
			singular:  "quote reason",
			header:    []string{columnID, columnCode, columnType, columnMessage},
			localized: true,
			row: func(r quoteReason, lang language) []string {
				return []string{r.ID, r.Code, r.Type, r.Message.in(lang)}
			},
		},
	}
	return newResourceCommand(a, reasons, commandsOf(a, quoteReasonWrites(reasons))...)
}
