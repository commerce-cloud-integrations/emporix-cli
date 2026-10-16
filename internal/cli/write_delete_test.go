package cli

import (
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

// confirmation returns answers to the confirmation of the case's write on a
// terminal, each with whether it confirms the write, and the question that
// follows the prompt's "Send METHOD URL to tenant acme1? ": [y/N] for a write
// but a delete, which y or yes confirms, and for a delete, which cannot be
// undone, what it deletes with one request, if more than one item (tt.many), and the
// word to type, tt.confirmBy, which alone confirms it.
func (tt *writeCase) confirmation(t *testing.T) (answers map[string]bool, question string) {
	t.Helper()
	if tt.method != http.MethodDelete {
		return map[string]bool{"n\n": false, "": false, "yes please\n": false, " Y \n": true, "YES": true}, "[y/N] "
	}
	if tt.confirmBy == "" {
		t.Fatalf("%q: a delete case states what confirms it (confirmBy)", tt.args)
	}
	deletes := "It cannot be undone"
	if tt.many != "" {
		deletes = "It deletes " + deletesMany(tt.many) + " and cannot be undone"
	}
	word := tt.confirmBy
	return map[string]bool{"y\n": false, "yes\n": false, "": false, "\n": false, word + "x\n": false,
		"x" + word + "\n": false, " " + word + " \n": true, word: true}, deletes + ": type " + word + " to confirm: "
}

// TestDeletesSayWhatConfirmsThem checks that the help of every delete says
// that it cannot be undone and what to type to confirm it: the name in its
// usage of the argument that its case types (confirmBy), or the tenant.
func TestDeletesSayWhatConfirmsThem(t *testing.T) {
	for _, tt := range writeCases() {
		if tt.method != http.MethodDelete {
			continue
		}
		cmd, rest, err := newTestRoot().Find(tt.args)
		if err == nil {
			err = cmd.ParseFlags(rest)
		}
		if err != nil {
			t.Fatalf("%q: %v", tt.args, err)
		}
		word := "the tenant"
		if i := slices.Index(cmd.Flags().Args(), tt.confirmBy); i >= 0 {
			word = strings.Fields(cmd.Use)[1+i]
		} else if tt.confirmBy != "acme1" {
			t.Fatalf("%q: confirmed by %q, which is neither an argument nor the tenant", tt.args, tt.confirmBy)
		}
		want := "which cannot be undone"
		if !strings.Contains(cmd.Long, want) {
			t.Errorf("%s: help %q; want %q", cmd.CommandPath(), cmd.Long, want)
		}
		if want := "\nOn a terminal, epx sends DELETE only once you type " + word + " at its prompt.\n"; !strings.Contains(
			cmd.Long, want) {
			t.Errorf("%s: help %q; want %q", cmd.CommandPath(), cmd.Long, want)
		}
	}
}

// TestDeletesOfManySaySo checks that the help of a delete of many items at
// once, always or with a flag, says what it deletes.
func TestDeletesOfManySaySo(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: []string{"units", "bulk-delete"},
			want: "Deletes the units whose codes --file names with one request, which cannot be undone."},
		{args: []string{"availability", "bulk-delete"},
			want: "Deletes the availability that --file names with one request, which cannot be undone."},
		{args: []string{"categories", "assignments", "delete-all"},
			want: "Deletes the assignments of every product to the category with one request, which cannot be undone."},
		{args: []string{"categories", "assignments", "delete-product-everywhere"},
			want: "Deletes the assignments of the product to every category with one request, which cannot be undone."},
		{args: []string{"categories", "delete"}, want: "Deletes data, which cannot be undone.\n" +
			"With --with-subcategories, it deletes the category and every category below it with one request."},
		{args: []string{"products", "delete"}, want: "Deletes data, which cannot be undone.\n" +
			"With --force, it deletes the product and every variant below it with one request."},
		{args: []string{"iam", "groups", "users", "delete-all"},
			want: "Deletes the membership of every user of the group with one request, which cannot be undone."},
		{args: []string{"iam", "users", "groups", "delete-all"},
			want: "Deletes the user's membership of every group with one request, which cannot be undone."},
		{args: []string{"iam", "groups", "delete"}, want: "Deletes data, which cannot be undone.\n" +
			"With --force-delete, it deletes the group and the membership of every user of it with one request."},
		{args: []string{"price-lists", "prices", "bulk-delete"},
			want: "Deletes the prices of a price list that --file names with one request, which cannot be undone."},
		{args: []string{"price-lists", "delete"},
			want: "Deletes the price list and every price in it with one request, which cannot be undone."},
		{args: []string{"price-models", "delete"}, want: "Deletes data, which cannot be undone.\n" +
			"With --force, it deletes the price model and every price that uses it with one request."},
		{args: []string{"fees", "assignments", "delete"},
			want: "Deletes every fee of the item YRN with one request, which cannot be undone."},
		{args: []string{"fees", "assignments", "delete-product"},
			want: "Deletes every fee of the product PRODUCT_ID with one request, which cannot be undone."},
		{args: []string{"units", "update"}, want: "Changes data: on a terminal, epx asks before it sends PUT."},
	} {
		if _, help, _ := runCLI(t, nil, append(tt.args, "--help")...); !strings.Contains(help, "\n\n"+tt.want) {
			t.Errorf("%q: help %q; want %q", tt.args, help, tt.want)
		}
	}
}

// partDelete is a delete whose request names two arguments and the site, and
// that deletes many items with one request.
func partDelete() *write {
	w := deleteWrite("delete-all THING_ID ID", "", "", request{path: "/things/{parent}/parts/{id}",
		query: url.Values{"siteCode": {"{site}"}}}, []string{"thing.manage"})
	w.deletes = deletion{many: "every part of the thing"}
	return w
}

// TestDeleteConfirmationNamesWhatItDeletes checks the prompt of a delete:
// it says what a delete of many items deletes and that it cannot be undone,
// and only the argument that fills its request's last placeholder, as given
// but for spaces around it, confirms it.
func TestDeleteConfirmationNamesWhatItDeletes(t *testing.T) {
	fake := newFakeEmporix(t)
	args := []string{"delete-all", "t1", "Part 1", "--site", "AcmeStore_NL"}
	prompt := "Send DELETE " + fake.server.URL + "/things/t1/parts/Part%201?siteCode=AcmeStore_NL to tenant acme1? " +
		"It deletes every part of the thing with one request and cannot be undone: type Part 1 to confirm: "
	for _, answer := range []string{"y\n", "t1\n", "part 1\n", "Part  1\n", "Part 1 x\n", "AcmeStore_NL\n"} {
		root := writeRootOn(fake, partDelete(), true)
		root.SetIn(strings.NewReader(answer))
		code, _, stderr := runWrite(t.Context(), root, args...)
		if code != ExitError {
			t.Errorf("answered %q: exit %d; want 1", answer, code)
		}
		if want := prompt + "Error: " + errNotConfirmed.Error() + "\n"; stderr != want {
			t.Errorf("answered %q: stderr %q; want %q", answer, stderr, want)
		}
	}
	fake.assertNothingSent(t)
	fake.route("/things/t1/parts/Part%201", respond(http.StatusNoContent, ""))
	root := writeRootOn(fake, partDelete(), true)
	root.SetIn(strings.NewReader("  Part 1 \n"))
	if code, _, stderr := runWrite(t.Context(), root, args...); code != ExitOK || stderr != prompt {
		t.Errorf("exit %d, stderr %q; want it confirmed after %q", code, stderr, prompt)
	}
	fake.assertRequests(t, "/things/t1/parts/Part%201?siteCode=AcmeStore_NL")
}

// TestDeleteOfNoItemIsConfirmedByTheTenant checks that a delete whose
// request names no argument, such as a bulk-delete, asks for the tenant, as
// its help says.
func TestDeleteOfNoItemIsConfirmedByTheTenant(t *testing.T) {
	fake := newFakeEmporix(t)
	tenantWide := deleteWrite(bulkDeleteUse, "", "", request{path: "/things"}, []string{"thing.manage"})
	tenantWide.deletes = deletion{many: "every thing"}
	if want := "only once you type the tenant at its prompt"; !strings.Contains(tenantWide.command(&app{}).Long, want) {
		t.Errorf("help %q; want %q", tenantWide.command(&app{}).Long, want)
	}
	for answer, code := range map[string]int{"bulk-delete\n": ExitError, "\n": ExitError, "acme1\n": ExitOK} {
		fake.route("/things", respond(http.StatusNoContent, ""))
		root := writeRootOn(fake, tenantWide, true)
		root.SetIn(strings.NewReader(answer))
		got, _, stderr := runWrite(t.Context(), root, bulkDeleteUse)
		if want := "It deletes every thing with one request and cannot be undone: type acme1 to confirm: "; got != code ||
			!strings.Contains(stderr, want) {
			t.Errorf("answered %q: exit %d, stderr %q; want exit %d after %q", answer, got, stderr, code, want)
		}
	}
	fake.assertRequests(t, "/things")
}

// TestDeleteConfirmationShowsWhatItCompares checks that a delete of an ID
// with a character that a terminal hides asks to type the ID as shown, which
// confirms it as the ID as given does, and that an ID shown as nothing but
// spaces cannot be confirmed by an empty answer.
func TestDeleteConfirmationShowsWhatItCompares(t *testing.T) {
	for _, tt := range []struct {
		id, answer string
		code       int
	}{
		{id: "a\u200bb", answer: "a b\n", code: ExitOK},
		{id: "a\u200bb", answer: "a\u200bb\n", code: ExitOK},
		{id: "a\u200bb", answer: "ab\n", code: ExitError},
		{id: "\u200b", answer: "\n", code: ExitError},
		{id: "\u200b", answer: "", code: ExitError},
	} {
		fake := newFakeEmporix(t)
		fake.route("/things/"+url.PathEscape(tt.id), respond(http.StatusNoContent, ""))
		root := writeRootOn(fake, deleteWrite("delete ID", "", "", request{path: "/things/{id}"},
			[]string{"thing.manage"}), true)
		root.SetIn(strings.NewReader(tt.answer))
		code, _, stderr := runWrite(t.Context(), root, "delete", tt.id)
		if code != tt.code || strings.ContainsRune(stderr, '\u200b') {
			t.Errorf("%q answered %q: exit %d, stderr %q; want exit %d and the ID as shown", tt.id, tt.answer, code,
				stderr, tt.code)
		}
	}
}

// TestBlankIDsAreRefused checks that an ID that is nothing but spaces, which
// addresses no item and could not be confirmed, is refused as a usage error
// before anything is sent.
func TestBlankIDsAreRefused(t *testing.T) {
	fake := newFakeEmporix(t)
	code, _, stderr := runIn(t, writeEnv(fake, false, ""), "units", "delete", " ", "--yes")
	if want := `invalid CODE " "`; code != ExitUsage || !strings.Contains(stderr, want) {
		t.Errorf("exit %d, stderr %q; want exit 2 and %q", code, stderr, want)
	}
	fake.assertNothingSent(t)
}

// TestDeletionsAreDeclaredOnDeletes checks that a write whose deletion is a
// mistake of its declaration panics while the command tree is built: one of
// a write that is not a delete, a cascade without what it deletes or without
// its switch, and a delete of no one item that does not say what it deletes;
// a cascade of a flag of the write's choice is declared (no panic).
func TestDeletionsAreDeclaredOnDeletes(t *testing.T) {
	force := []paramFlag{{param: "force", flag: "force"}}
	forceChoice := queryChoice{param: "force", options: []choiceOption{{flag: "force", value: queryTrue,
		scopes: []string{"thing.part_manage"}}}}
	for _, tt := range []struct {
		method   string
		req      request
		switches []paramFlag
		choice   queryChoice
		deletes  deletion
		want     string
	}{
		{method: http.MethodPut, req: request{path: "/things"}, deletes: deletion{many: "every thing"},
			want: "x: only a delete says what it deletes"},
		{method: http.MethodDelete, req: request{path: "/things/{id}"}, deletes: deletion{many: "parts", with: "force"},
			want: "x: a delete deletes many with a switch of its own: what and which"},
		{method: http.MethodDelete, req: request{path: "/things/{id}"}, switches: force,
			deletes: deletion{with: "force"}, want: "x: a delete deletes many with a switch of its own: what and which"},
		{method: http.MethodDelete, req: request{path: "/things/{parent}/parts"},
			want: "x: a delete of no {id} says what it deletes"},
		{method: http.MethodDelete, req: request{path: "/things/{id}"}, choice: forceChoice,
			deletes: deletion{many: "the thing and its parts", with: "force"}},
	} {
		func() {
			defer func() {
				if recovered, _ := recover().(string); recovered != tt.want {
					t.Errorf("%s %s: recovered %q; want %q", tt.method, tt.req.path, recovered, tt.want)
				}
			}()
			(&write{use: "x", method: tt.method, req: tt.req, scopes: []string{"thing.manage"}, body: noBody{},
				switches: tt.switches, choice: tt.choice, deletes: tt.deletes}).command(&app{})
		}()
	}
}

// TestConfirmationFailsWhereItCannotAsk checks that a write, a delete or
// another, whose question cannot be written on standard error fails (exit 1)
// and sends nothing, whatever standard input answers.
func TestConfirmationFailsWhereItCannotAsk(t *testing.T) {
	for _, method := range []string{"delete", "post"} {
		fake := newFakeEmporix(t)
		env := writeEnv(fake, true, "/x\ny\n")
		env.Out, env.Err = io.Discard, failingWriter{}
		if code := Run(t.Context(), []string{"api", method, "/x"}, env); code != ExitError {
			t.Errorf("api %s: exit %d; want 1", method, code)
		}
		fake.assertRequests(t)
	}
}
