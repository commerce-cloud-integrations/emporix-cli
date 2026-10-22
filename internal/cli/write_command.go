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

// write declares one write command of a group: what it sends, where, the
// scopes it requests and the job it starts, if any. Every write command but
// those of api comes from one (command), sent through a sender, so none
// sends otherwise or requests a scope that it does not declare.
type write struct {
	use      string      // name and arguments: {parent}, {inner} and {id} in req, then the body's, e.g. "update KEY"
	short    string      // summary
	long     string      // what it sends and what Emporix answers; how it is sent and the scope sentence follow
	example  string      // command lines that show it, each indented by two spaces; none when empty
	method   string      // http.MethodPost, MethodPut, MethodPatch or MethodDelete
	req      request     // where it sends: a path and query that may name {parent}, {inner}, {id} and {site}
	scopes   []string    // OAuth scopes of the write besides the tenant: a write-capable one at least, unless readScoped
	body     bodySource  // where its body comes from: --file or typed flags (a payload of its own), its arguments or none
	params   queryFlags  // query parameters that flags set, each sent when its flag is given
	headers  []paramFlag // headers that flags set, each sent when its flag is given, e.g. the session of a chat
	switches []paramFlag // query parameters sent as true with their flag and false without (no choices)
	choice   queryChoice // flags that each add scopes and may set a query parameter; none without options
	// paramScopes is a query parameter, besides params, whose flag sets it,
	// with scopes that it requests too only where its query needs them; none
	// without scopes.
	paramScopes paramScopes
	deletes     deletion // of a delete: what it deletes besides the item that it names, as its confirmation says
	// language says whether its body holds localized text and how it tells
	// Emporix the language of that text (languageMode); none when not given.
	language languageMode
	// answers says how Emporix answers it beyond its status (emporix.Answers):
	// a stale version or a refused input, an outcome that fails the run unless
	// true, or a conflict that is answered with a body that says why.
	answers emporix.Answers
	// choices holds the only values that an argument of req takes, by its
	// name in use, e.g. TYPE: PRODUCT and CATEGORY; any when nil.
	choices map[string][]string
	// predefined holds values that an argument takes besides others, by its
	// name in use; they complete as its choices do (command.predefined).
	predefined map[string][]string
	// answer is how a run ends once Emporix answered it (answerer), such as
	// following the job that it starts (--wait); by printing the response
	// when nil.
	answer   answerer
	conflict string // what a 409 Conflict means and how to recover; that the item changed when empty (finish)
	// languageHelp is what the help of a write of stringsOrMaps or mapsOnly
	// says of the localized text of its body; contentLanguageHelp or
	// mapsOnlyHelp when empty. headerlessMaps has a help of its own
	// (headerlessMapsHelp), and notLocalized none.
	languageHelp string
	ids          argumentIDs // the reads that list the IDs of its arguments, for completion
	// reserved holds the values of the ID argument ({id}, which req must name)
	// that address another operation of the collection rather than one item,
	// each with the command that sends that operation, e.g. bulk: bulk-update.
	reserved map[string]string
	// readScoped says that Emporix carries out the write without a
	// write-capable scope: with scopes named for reading, which scopes then
	// holds, and none other, or with none at all. It is an action that changes
	// data although Emporix asks no write scope for it, such as issuing the
	// next ID of a sequence, which advances its counter, or creating a
	// customer's referral coupon. Its help says so (readScopedNote); its scope
	// sentence names the scopes, if any.
	readScoped bool
}

// command returns the write command of w. It takes the arguments of w.req and
// its --site as a read takes those of its request (request.bind), then those of
// its body, the flags of every write (sender), the body it sends,
// --content-language where its language lets it name one (bindLanguage),
// sent as its Content-Language (languageHeader), the flags of its
// query parameters and headers, a flag per query switch and those of its
// choice; it sends with a token of w's scopes, with those of its parameter
// where its query needs them (paramScopes), those of the custom entity type
// that its first argument names where they name one (typeScopes), and those
// of the choice's flag given, if any, after any read of its body with the
// body's own, and says which scopes in its help and its annotation, the body's
// first (the help names those of the choice's flags and of its parameter
// apart); it takes the flags
// of its answer, if any: --wait, which follows the job that it starts with the
// scopes of reading it too (jobWait); and it is confirmed as w.confirmation
// says. A write without a write-capable scope that is not declared readScoped,
// one declared so with a write-capable scope, one whose choice sets a
// parameter that its request sends already, one that reserves values of an ID
// it does not take, one with a languageHelp that would not say how it sends
// its localized text (bindLanguage), or one whose deletion is not that of a delete
// (checkDeletion) is a mistake in its declaration (the first would be sent
// with scopes that cannot change data), so it panics while the command tree
// is built, which every test of the tree does.
func (w *write) command(a *app) *cobra.Command {
	w.checkDeclaration()
	requested := w.requested()
	cmd := &cobra.Command{
		Use: w.use, Short: w.short, Long: w.what(), Example: w.example,
		Annotations: map[string]string{writeScopesAnnotation: strings.Join(requested, " ")},
	}
	if w.answers.ForbiddenIsRefusal {
		checkRefusalHelp(cmd)
	}
	site := w.bindArguments(a, cmd)
	if len(w.reserved) > 0 {
		if !w.req.names(placeholderID) {
			panic(w.use + ": a write reserves values of an ID only where its request names " + placeholderID)
		}
		cmd.Args = refuseReserved(cmd.Args, w.reserved)
	}
	// after bind and refuseReserved, whose argument rules it wraps; every scope that it may request
	fillType := typeScopes(cmd, w.req, slices.Concat(w.scopes, w.paramScopes.scopes))
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
	cmd.Long += w.choice.scopeNote() + w.paramScopes.note() + "\n\n" + writeScopeSentence(requested)
	n := len(w.req.arguments())
	cmd.RunE = func(c *cobra.Command, args []string) error {
		path, query := w.req.fill(args, *site)
		setParams(query)
		needed := paramScoped(query)
		for _, sw := range w.switches {
			// The flag was defined above as a bool, so reading it cannot fail.
			on, _ := c.Flags().GetBool(sw.flag)
			query.Set(sw.param, strconv.FormatBool(on))
		}
		sent := &emporix.Write{Path: path, Query: query, Answers: w.answers, Header: http.Header{}}
		if header := w.languageHeader(*lang); header != "" {
			sent.Header.Set("Content-Language", header)
		}
		setHeaders(sent.Header)
		scopes, then := end.plan(c, choose(query, fillType(args, slices.Concat(w.scopes, needed))))
		return s.sendThen(c, scopes, sent, args[n:], w.subject(c, args), then)
	}
	return cmd
}

// bindArguments binds the arguments of w's request, and --site where it
// names the site, to cmd, which completes them with their choices and
// predefined values and as w.ids lists them (completeArguments), and
// returns the site that --site gives.
func (w *write) bindArguments(a *app, cmd *cobra.Command) *siteCode {
	site := w.req.bind(cmd, w.choices)
	completeArguments(a, cmd, site, w.ids, w.choices, w.predefined)
	return site
}

// requested returns the scopes that w's help and annotation name: those of
// any read of its body, then its own. Only its own may name a custom entity
// type, which typeScopes fills in (refuseTypeScopes).
func (w *write) requested() []string {
	refuseTypeScopes(w.use, w.body.scopes())
	return slices.Concat(w.body.scopes(), w.scopes)
}

// readScopedNote ends what a readScoped write sends, in its help.
const readScopedNote = "It changes data although Emporix needs no write-capable scope for it."

// checkDeclaration panics where w is a mistake of its declaration, as
// command says: one without a write-capable scope that is not declared
// readScoped, one declared so with one, one whose choice sets a parameter
// that its request sends already, or one whose deletion is not that of a
// delete (checkDeletion).
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

// what returns what the help of w says that it sends: long, followed by
// readScopedNote for a readScoped write.
func (w *write) what() string {
	if w.readScoped {
		return w.long + "\n" + readScopedNote
	}
	return w.long
}

// confirmation returns how w is confirmed, as its method asks
// (confirmationFor): a delete by typing the argument that fills the last
// placeholder of its request, the item that it addresses, as its usage names
// it, or the tenant where it takes none.
func (w *write) confirmation() confirmation {
	word := "the tenant"
	if n := len(w.req.arguments()); n > 0 {
		word = argumentNames(w.use, n)[n-1]
	}
	return confirmationFor(w.method, word, w.deletes)
}

// subject returns what the confirmation of w's run on cmd with args names:
// the argument that fills the last placeholder of its request, if any, and
// what it deletes with the same request besides.
func (w *write) subject(cmd *cobra.Command, args []string) subject {
	about := subject{many: w.deletes.at(cmd)}
	if n := len(w.req.arguments()); n > 0 {
		about.word = args[n-1]
	}
	return about
}

// deletion says what a delete deletes besides the item that its arguments
// name, which its help and its prompt say: nothing (the zero value: one
// item), many, what it deletes with one request, such as "the units whose
// codes --file names", or that only with, the flag of one of its switches or of
// its choice (a cascade, such as --force).
type deletion struct {
	many string
	with string
}

// sentence says, in the help of a delete, that it cannot be undone and what
// it deletes with one request (deletesMany).
func (d deletion) sentence() string {
	switch {
	case d.many == "":
		return "Deletes data, which cannot be undone."
	case d.with == "":
		return "Deletes " + deletesMany(d.many) + ", which cannot be undone."
	default:
		return "Deletes data, which cannot be undone.\nWith --" + d.with + ", it deletes " + deletesMany(d.many) + "."
	}
}

// deletesMany says what a delete deletes, many (deletion.many), as its help
// and its prompt say it: with one request, which Emporix may carry out in
// the background.
func deletesMany(many string) string { return many + " with one request" }

// at returns what the delete of cmd, once its flags are parsed, deletes with
// the same request besides one item: many, unless the flag of with is off; none for
// one item.
func (d deletion) at(cmd *cobra.Command) string {
	if d.with != "" {
		// with names a switch or choice flag of the write, a bool flag (checkDeletion), so reading it cannot fail.
		if on, _ := cmd.Flags().GetBool(d.with); !on {
			return ""
		}
	}
	return d.many
}

// checkDeletion panics where w's deletion is a mistake of its declaration:
// one of a write that is not a delete, a cascade without many or without the
// switch or choice flag that it names, or a delete of no {id} (of no one
// item, such as a bulk-delete) that does not say what it deletes.
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

// deleteWrite declares the delete use, a DELETE to req without a body, with
// scopes; long says what Emporix deletes.
func deleteWrite(use, short, long string, req request, scopes []string) *write {
	return &write{use: use, short: short, long: long, method: http.MethodDelete, req: req, scopes: scopes, body: noBody{}}
}

// ifUnchangedNote says how a write that sends metadata.version avoids
// overwriting a change made since the item was read.
const ifUnchangedNote = "Send its metadata.version as read to change it only if nobody changed it since:\n" +
	"Emporix answers 409 Conflict (exit 5) otherwise."

// unknownAsDeleted ends the help of a delete that Emporix answers for an
// unknown item, named by the argument arg, as for one that it deleted.
func unknownAsDeleted(arg string) string {
	return "\nEmporix answers an unknown " + arg + " as one that it deleted (204 No Content)."
}

// itemDelete declares `delete ID` of r: a DELETE of the item ID (its
// itemPath), as deleteWrite declares one, whose ID completes from r's list
// (itemIDs).
func itemDelete[T any](r resource[T], short, long string, scopes []string) *write {
	w := deleteWrite(r.itemUse(deleteUse), short, long, request{path: r.itemPath()}, scopes)
	w.ids = r.itemIDs()
	return w
}

// commandsOf returns the write commands of writes, in their order.
func commandsOf(a *app, writes []*write) []*cobra.Command {
	commands := make([]*cobra.Command, len(writes))
	for i, w := range writes {
		commands[i] = w.command(a)
	}
	return commands
}

// refuseReserved returns accept, which takes the arguments of a command whose
// last is an ID, refusing also an ID that reserved names: one that addresses
// another operation of the collection, which the command would send to.
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

// bulkSegment is the path segment, below a collection, of its bulk writes.
const bulkSegment = "bulk"

// The names of the bulk writes of a collection: a JSON array of items, or of
// what names the items to delete, sent to its bulk path (bulkWrites), or to
// the collection.
const (
	bulkCreateUse = "bulk-create"
	bulkUpdateUse = "bulk-update"
	bulkPatchUse  = "bulk-patch"
	bulkDeleteUse = "bulk-delete"
)

// bulkWrites declares the bulk writes of a collection, those of bulk-create
// (POST), bulk-update (PUT, an upsert), bulk-patch (PATCH) and bulk-delete
// (DELETE) whose help it gives: a JSON array of items, or of what names the
// items to delete, sent to the bulk path of the collection, which Emporix
// carries out item by item (sender), or, for a bulk-delete that deleteWhole
// marks, answers as a whole. A collection below an item of another, such as
// the customers of a segment, takes the arguments that name that item.
type bulkWrites struct {
	path     string              // the collection's path, which may name {parent} and {inner}; they send below it
	args     string              // the arguments that fill in path, e.g. "SEGMENT_ID TYPE"; none for the tenant's
	nouns    string              // the items in their summaries, e.g. "prices"
	scopes   []string            // the write scopes besides the tenant; one must be write-capable
	maxItems int                 // the most items that Emporix takes in one request, as its reference documents
	choices  map[string][]string // the only values of an argument, by its name in args; any when nil
	create   string              // what bulk-create sends; how it is sent and the scope sentence follow; none when empty
	update   string              // what bulk-update sends, alike
	patch    string              // what bulk-patch sends, alike
	delete   string              // what bulk-delete deletes, alike
	// language says whether the items of bulk-create and bulk-update hold
	// localized text, and how they tell Emporix its language (languageMode).
	language languageMode
	// deleteSwitches are the query switches of bulk-delete (write.switches).
	deleteSwitches []paramFlag
	// deleteWhole says that Emporix answers bulk-delete as a whole (e.g. 204
	// No Content), not item by item; its items are still counted.
	deleteWhole bool
	// ids are the reads that list the IDs of args, for completion.
	ids argumentIDs
}

// bulkOperation is one bulk write of a collection: its name, the verb of its
// summary, its method and its help, as bulkWrites declares it.
type bulkOperation struct{ use, does, method, long string }

// operations returns the bulk writes that b declares, those whose help it
// gives, in this order: bulk-create, bulk-update, bulk-patch, bulk-delete.
func (b *bulkWrites) operations() []bulkOperation {
	return slices.DeleteFunc([]bulkOperation{
		{bulkCreateUse, "Create", http.MethodPost, b.create},
		{bulkUpdateUse, "Upsert", http.MethodPut, b.update},
		{bulkPatchUse, "Change", http.MethodPatch, b.patch},
		{bulkDeleteUse, "Delete", http.MethodDelete, b.delete},
	}, func(op bulkOperation) bool { return op.long == "" })
}

// commands returns the commands of b's writes.
func (b *bulkWrites) commands(a *app) []*cobra.Command { return writeCommands(a, b.writes()) }

// writes declares the bulk writes of b (operations), each with a body of its
// own and taking args, if any. A caller may complete them, such as with query
// parameters, before it makes their commands.
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

// reserved returns what write.reserved holds for a write of one item of b's
// collection, sent as method to path: bulk, with the bulk write that it would
// send, where path names the item as {id} right below the collection and b
// declares a bulk write sent as method; none otherwise, also for no b.
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

// writeCommands returns the command of each of writes, in order.
func writeCommands(a *app, writes []write) []*cobra.Command {
	commands := make([]*cobra.Command, len(writes))
	for i := range writes {
		commands[i] = writes[i].command(a)
	}
	return commands
}

// forceParam is the query parameter, and the flag, of a write that Emporix
// carries out only when forced: a delete of an item still in use, or the
// disabling or deleting of an item that others use.
const forceParam = "force"

// partialSwitch is the --partial of a write that sends partial=true to
// change only the fields its body gives, and partial=false otherwise, to
// replace the whole of what, e.g. "price".
func partialSwitch(what string) paramFlag {
	return paramFlag{param: partialParam, flag: partialParam,
		usage: "change only the fields that the body gives; without it, the body replaces the whole " + what}
}

// outcomeHelp follows what a write whose response says whether it was carried
// out (write.answers.ReportsOutcome) sends, in its help.
const outcomeHelp = "Emporix answers whether the operation succeeded (successful). epx prints the response and\n" +
	"exits with 1 unless it says that it did: where it says that it failed (successful false,\n" +
	"with a message), and where it does not say, when it may have been carried out all the same."
