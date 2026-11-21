package cli

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type write struct {
	use          string
	short        string
	long         string
	example      string
	method       string
	req          request
	scopes       []string
	body         bodySource
	params       queryFlags
	headers      []paramFlag
	switches     []paramFlag
	choice       queryChoice
	paramScopes  paramScopes
	deletes      deletion
	language     languageMode
	answers      emporix.Answers
	choices      map[string][]string
	builtIn      builtInTypes
	answer       answerer
	conflict     string
	languageHelp string
	ids          argumentIDs
	reserved     map[string]string
	readScoped   bool
}

func (w *write) command(a *app) *cobra.Command {
	w.checkDeclaration()
	scopes := w.requested()
	cmd := &cobra.Command{
		Use: w.use, Short: w.short, Long: w.what(), Example: w.example,
		Annotations: map[string]string{writeScopesAnnotation: strings.Join(scopes, " ")},
	}
	if w.answers.ForbiddenIsRefusal {
		checkRefusalHelp(cmd)
	}
	site, rule := w.bindArguments(a, cmd)
	lang := w.bindLanguage(cmd)
	if w.answers.ReportsOutcome {
		cmd.Long += "\n\n" + outcomeHelp
	}
	s := &sender{a: a, method: w.method, body: w.body, confirm: w.confirmation(),
		conflict: w.conflict}
	s.bind(cmd)
	setParams := w.params.bind(cmd)
	paramScoped := w.paramScopes.bind(cmd)
	setHeaders := queryFlags{optional: w.headers}.bind(cmd)
	for _, sw := range w.switches {
		cmd.Flags().Bool(sw.flag, false, sw.usage)
	}
	choose := w.choice.bind(cmd)
	var end endingPlanner = printing{}
	if w.answer != nil {
		end = w.answer.bind(a, cmd)
	}
	long := joinParagraphs(cmd.Long, w.choice.scopeNote(), w.paramScopes.note())
	cmd.Long = rule.helpEnding(long, writeScopeSentence(scopes))
	n := len(w.req.arguments())
	cmd.RunE = func(c *cobra.Command, args []string) error {
		path, query := w.req.fill(args, *site)
		setParams(query)
		needed := paramScoped(query)
		for _, sw := range w.switches {
			on, _ := c.Flags().GetBool(sw.flag)
			query.Set(sw.param, strconv.FormatBool(on))
		}
		sent := &emporix.Write{Path: path, Query: query, Answers: w.answers, Header: http.Header{}}
		if header := w.languageHeader(*lang); header != "" {
			sent.Header.Set("Content-Language", header)
		}
		setHeaders(sent.Header)
		scopes, then := end.plan(c, choose(query, rule.fill(args, slices.Concat(w.scopes, needed))))
		return s.sendThen(c, scopes, sent, args[n:], w.subject(c, args), then)
	}
	return cmd
}

func (w *write) bindArguments(a *app, cmd *cobra.Command) (site *siteCode, rule typeRule) {
	site = w.req.bind(cmd, w.choices)
	if len(w.reserved) > 0 {
		if !w.req.names(placeholderID) {
			panic(w.use + ": a write reserves values of an ID only where its request names " + placeholderID)
		}
		cmd.Args = refuseReserved(cmd.Args, w.reserved)
	}
	rule = typeScopes(cmd, w.req, slices.Concat(w.scopes, w.paramScopes.scopes), w.builtIn)
	completeArguments(a, cmd, site, w.ids, w.choices, rule)
	return site, rule
}

func (w *write) requested() []string {
	refuseTypeScopes(w.use, w.body.scopes())
	return slices.Concat(w.body.scopes(), w.scopes)
}

const readScopedNote = "It changes data although Emporix needs no write-capable scope for it."

func (w *write) checkDeclaration() {
	switch capable := writeCapable(w.scopes); {
	case w.readScoped && capable:
		panic(w.use + ": a write declared readScoped requests read scopes only")
	case !w.readScoped && !capable:
		panic(fmt.Sprintf("%s: a write requests a write-capable scope", w.use))
	case w.req.query.Has(w.choice.param):
		panic(fmt.Sprintf("%s: its choice sets %s, which its request sends already", w.use, w.choice.param))
	}
	w.checkDeletion()
}

func (w *write) what() string {
	if w.readScoped {
		return joinSentences(w.long, readScopedNote)
	}
	return w.long
}

func (w *write) confirmation() confirmation {
	word := "the tenant"
	if n := len(w.req.arguments()); n > 0 {
		word = argumentNames(w.use, n)[n-1]
	}
	return confirmationFor(w.method, word, w.deletes)
}

func (w *write) subject(cmd *cobra.Command, args []string) subject {
	about := subject{many: w.deletes.at(cmd)}
	if n := len(w.req.arguments()); n > 0 {
		about.word = args[n-1]
	}
	return about
}

type deletion struct {
	many string
	with string
}

func (d deletion) sentence() string {
	switch {
	case d.many == "":
		return "Deletes data, which cannot be undone."
	case d.with == "":
		return "Deletes " + deletesMany(d.many) + ", which cannot be undone."
	default:
		return "Deletes data, which cannot be undone. With --" + d.with + ", it deletes " + deletesMany(d.many) + "."
	}
}

func deletesMany(many string) string { return many + " with one request" }

func (d deletion) at(cmd *cobra.Command) string {
	if d.with != "" {
		if on, _ := cmd.Flags().GetBool(d.with); !on {
			return ""
		}
	}
	return d.many
}

func (w *write) checkDeletion() {
	d := w.deletes
	hasSwitch := slices.ContainsFunc(w.switches, func(sw paramFlag) bool { return sw.flag == d.with }) ||
		slices.ContainsFunc(w.choice.options, func(option choiceOption) bool { return option.flag == d.with })
	switch {
	case w.method != http.MethodDelete && d != (deletion{}):
		panic(w.use + ": only a delete says what it deletes")
	case d.with != "" && (d.many == "" || !hasSwitch):
		panic(w.use + ": a delete deletes many with a switch of its own: what and which")
	case w.method == http.MethodDelete && !w.req.names(placeholderID) && d.many == "":
		panic(w.use + ": a delete of no " + placeholderID + " says what it deletes")
	}
}

func deleteWrite(use, short, long string, req request, scopes []string) *write {
	return &write{use: use, short: short, long: long, method: http.MethodDelete, req: req, scopes: scopes, body: noBody{}}
}

const ifUnchangedNote = "Send its metadata.version as read to change it only if nobody changed it since: " +
	"Emporix answers 409 Conflict (exit 5) otherwise."

func unknownAsDeleted(arg string) string {
	return "Emporix answers an unknown " + arg + " as one that it deleted (204 No Content)."
}

func itemDelete[T any](r resource[T], short, long string, scopes []string) *write {
	w := deleteWrite(r.itemUse(deleteUse), short, long, request{path: r.itemPath()}, scopes)
	w.ids = r.itemIDs()
	return w
}

func commandsOf(a *app, writes []*write) []*cobra.Command {
	commands := make([]*cobra.Command, len(writes))
	for i, w := range writes {
		commands[i] = w.command(a)
	}
	return commands
}

func refuseReserved(accept cobra.PositionalArgs, reserved map[string]string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := accept(cmd, args); err != nil {
			return err
		}
		last := len(args) - 1
		if other, isReserved := reserved[args[last]]; isReserved {
			return fmt.Errorf("invalid %s %q: it would be sent as %s, not to one item",
				argumentNames(cmd.Use, len(args))[last], args[last], other)
		}
		return nil
	}
}

const bulkSegment = "bulk"

const (
	bulkCreateUse = "bulk-create"
	bulkUpdateUse = "bulk-update"
	bulkPatchUse  = "bulk-patch"
	bulkDeleteUse = "bulk-delete"
)

type bulkWrites struct {
	path           string
	args           string
	nouns          string
	scopes         []string
	maxItems       int
	choices        map[string][]string
	create         string
	update         string
	patch          string
	delete         string
	language       languageMode
	deleteSwitches []paramFlag
	deleteWhole    bool
	ids            argumentIDs
}

type bulkOperation struct{ use, does, method, long string }

func (b *bulkWrites) operations() []bulkOperation {
	return slices.DeleteFunc([]bulkOperation{
		{bulkCreateUse, "Create", http.MethodPost, b.create},
		{bulkUpdateUse, "Upsert", http.MethodPut, b.update},
		{bulkPatchUse, "Change", http.MethodPatch, b.patch},
		{bulkDeleteUse, "Delete", http.MethodDelete, b.delete},
	}, func(op bulkOperation) bool { return op.long == "" })
}

func (b *bulkWrites) commands(a *app) []*cobra.Command { return writeCommands(a, b.writes()) }

func (b *bulkWrites) writes() []write {
	req := request{path: b.path + "/" + bulkSegment}
	var writes []write
	for _, op := range b.operations() {
		w := write{
			use: strings.TrimSpace(op.use + " " + b.args), short: op.does + " " + b.nouns + " in bulk",
			long: op.long, method: op.method, req: req, scopes: b.scopes, body: &payload{maxItems: b.maxItems},
			choices: b.choices, ids: b.ids,
		}
		switch op.method {
		case http.MethodPost, http.MethodPut:
			w.language = b.language
		case http.MethodDelete:
			w.switches = b.deleteSwitches
			w.deletes = deletion{many: "the " + b.nouns + " that --file names"}
			w.body = &payload{maxItems: b.maxItems, whole: b.deleteWhole}
		}
		writes = append(writes, w)
	}
	return writes
}

func (b *bulkWrites) reserved(method, path string) map[string]string {
	if b == nil || path != b.path+"/"+placeholderID {
		return nil
	}
	for _, op := range b.operations() {
		if op.method == method {
			return map[string]string{bulkSegment: op.use}
		}
	}
	return nil
}

func writeCommands(a *app, writes []write) []*cobra.Command {
	commands := make([]*cobra.Command, len(writes))
	for i := range writes {
		commands[i] = writes[i].command(a)
	}
	return commands
}

const forceParam = "force"

func partialSwitch(what string) paramFlag {
	return paramFlag{param: partialParam, flag: partialParam,
		usage: "change only the fields that the body gives; without it, the body replaces the whole " + what}
}

const outcomeHelp = "Emporix answers whether the operation succeeded (successful). epx prints the response and " +
	"exits with 1 unless it says that it did: where it says that it failed (successful false, " +
	"with a message), and where it does not say, when it may have been carried out all the same."
