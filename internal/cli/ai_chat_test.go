package cli

import (
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestAIChatsSendTheirSession checks that a chat and an attachment send the
// session that --session names as the session-id header, which a dry run
// shows, and none without it.
func TestAIChatsSendTheirSession(t *testing.T) {
	for _, tt := range []struct {
		args []string
		path string
	}{
		{args: []string{"ai", "chat", "--agent", "a1", "--message", "Hi"}, path: aiAgenticPath + "chat"},
		{args: []string{"ai", "chat-async", "--agent", "a1", "--message", "Hi"}, path: aiAgenticPath + "chat-async"},
		{args: []string{"ai", "agents", "attach", "a1", "--media", "m1"}, path: aiAgenticPath + "a1/attachments"},
	} {
		for session, want := range map[string]string{"s 1": "s 1", "": ""} {
			args := slices.Clone(tt.args)
			if session != "" {
				args = append(args, "--session", session)
			}
			fake := newFakeEmporix(t)
			var sent []string
			fake.route(tt.path, func(w http.ResponseWriter, r *http.Request) {
				sent = append(sent, r.Header.Get("session-id"))
				respond(http.StatusOK, `{"jobId":"j1"}`)(w, r)
			})
			code, _, stderr := runIn(t, writeEnv(fake, false, ""), append(args, "--yes")...)
			if code != ExitOK || !slices.Equal(sent, []string{want}) {
				t.Errorf("%q: exit %d, stderr %q, session-id %q; want %q", args, code, stderr, sent, want)
			}
			_, stdout, _ := runIn(t, writeEnv(fake, false, ""), append(args, "--dry-run")...)
			if strings.Contains(stdout, "\nSession-Id: s 1\n") != (want != "") {
				t.Errorf("%q --dry-run: %q; want the session shown exactly when given", args, stdout)
			}
		}
	}
}

// TestAIChatSendsAttachmentsOfItsSession checks that a chat sends each
// --attachment, in their order, with the session that they belong to.
func TestAIChatSendsAttachmentsOfItsSession(t *testing.T) {
	fake := newFakeEmporix(t)
	var session string
	fake.route(aiAgenticPath+"chat", func(w http.ResponseWriter, r *http.Request) {
		session = r.Header.Get("session-id")
		respond(http.StatusOK, `[]`)(w, r)
	})
	code, _, stderr := runIn(t, writeEnv(fake, false, ""), "ai", "chat", "--agent", "a1", "--message", "Order this",
		"--session", "s1", "--attachment", "att1", "--attachment", "att2", "--yes")
	want := `{"agentId":"a1","attachments":[{"attachmentId":"att1"},{"attachmentId":"att2"}],"message":"Order this"}`
	if code != ExitOK || session != "s1" || len(fake.bodies) != 1 || fake.bodies[0] != want {
		t.Errorf("exit %d, stderr %q, session %q, bodies %q; want %s in session s1", code, stderr, session,
			fake.bodies, want)
	}
}

// TestAIWriteUsageErrorsSendNothing checks that an attachment without a file
// or a media asset, or with both, with a file that cannot be read or comes
// from standard input, which has no name to send, a chat or generation
// without what it sends or with a message, attachment or number of tokens it
// cannot send, an attachment without the session it belongs to, and a
// session that cannot be sent as a header are usage errors (exit 2), also for
// a dry run, that send nothing.
func TestAIWriteUsageErrorsSendNothing(t *testing.T) {
	directory := t.TempDir()
	attach := []string{"ai", "agents", "attach", "a1"}
	chat := []string{"ai", "chat", "--agent", "a1", "--message", "Hi"}
	generate := []string{"ai", "generate-text", "--prompt", "Hi", "--max-tokens"}
	const badSession = "invalid argument \"s1\\r\\nX-Other: 1\" for \"--session\" flag: holds a control character"
	const notACount = "for \"--max-tokens\" flag: use a whole number above 0"
	for _, tt := range []struct {
		args []string
		want string
	}{
		{args: attach, want: "at least one of the flags in the group [upload media] is required"},
		{args: append(slices.Clone(attach), "--upload", "testdata/order.csv", "--media", "m1"),
			want: "if any flags in the group [upload media] are set none of the others can be"},
		{args: append(slices.Clone(attach), "--upload", "-"),
			want: `--upload "-": give the path of the file, whose name is sent with it`},
		{args: append(slices.Clone(attach), "--upload", directory), want: `": is a directory`},
		{args: append(slices.Clone(attach), "--upload", filepath.Join(directory, "none.pdf")), want: "no such file"},
		{args: append(slices.Clone(attach), "--media", ""), want: `invalid argument "" for "--media" flag: empty`},
		{args: []string{"ai", "chat", "--message", "Hi"}, want: "the flags in the group [file agent] is required"},
		{args: []string{"ai", "chat-async", "--agent", "a1"}, want: "the flags in the group [file message] is required"},
		{args: append(slices.Clone(chat), "--attachment", ""), want: `invalid argument "" for "--attachment" flag: empty`},
		{args: append(slices.Clone(chat), "--attachment", "att1"), want: "--attachment applies only with --session"},
		{args: append(slices.Clone(chat), "--session", "s1\r\nX-Other: 1"), want: badSession},
		{args: []string{"ai", "chat-async", "--agent", "a1", "--message", "Hi", "--session", "s1\r\nX-Other: 1"},
			want: badSession},
		{args: append(slices.Clone(attach), "--media", "m1", "--session", "s1\r\nX-Other: 1"), want: badSession},
		{args: []string{"ai", "generate-text", "--max-tokens", "500"},
			want: "the flags in the group [file prompt] is required"},
		{args: append(slices.Clone(generate), "many"), want: notACount},
		{args: append(slices.Clone(generate), "1.5"), want: notACount},
		{args: append(slices.Clone(generate), "0"), want: notACount},
		{args: append(slices.Clone(generate), "99999999999999999999"), want: notACount},
		{args: []string{"ai", "generate-completion"}, want: "the flags in the group [file message] is required"},
		{args: []string{"ai", "agents", "export"}, want: "the flags in the group [file agent] is required"},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			fake := newFakeEmporix(t)
			code, stdout, stderr := runIn(t, writeEnv(fake, false, ""), append(slices.Clone(tt.args), flag)...)
			if code != ExitUsage || stdout != "" || !strings.Contains(stderr, tt.want) {
				t.Errorf("%q %s: exit %d, stdout %q, stderr %q; want exit 2 and %q", tt.args, flag, code, stdout,
					stderr, tt.want)
			}
			fake.assertNothingSent(t)
		}
	}
	for _, message := range []string{"Describe product 100200", "USER:", "USER:  ", "AGENT:Hi", ":Hi"} {
		fake := newFakeEmporix(t)
		code, _, stderr := runIn(t, writeEnv(fake, false, ""), "ai", "generate-completion", "--message", message,
			"--dry-run")
		if code != ExitUsage || !strings.Contains(stderr, "use ROLE:TEXT, with ROLE SYSTEM, USER or ASSISTANT") {
			t.Errorf("--message %q: exit %d, stderr %q; want exit 2", message, code, stderr)
		}
		fake.assertNothingSent(t)
	}
}

// TestAIHelpNamesCreditsAndJobs checks that what consumes AI credits says so
// in its summary and help, that what waits for a reply names --timeout, and
// that the export says that it creates a job, although it needs a read scope
// only.
func TestAIHelpNamesCreditsAndJobs(t *testing.T) {
	root := newTestRoot()
	for _, use := range []string{"chat", "chat-async", "generate-text", "generate-completion"} {
		cmd, _, err := root.Find([]string{"ai", use})
		if err != nil || !strings.HasSuffix(cmd.Short, "(consumes AI credits)") ||
			!strings.Contains(cmd.Long, "It consumes AI credits: the LLM tokens of") {
			t.Errorf("ai %s: %v, summary %q, help %q; want both to name AI credits", use, err, cmd.Short, cmd.Long)
		}
		if strings.Contains(cmd.Long, "A reply may take longer than the --timeout of a request (30s by "+
			"default)") == (use == "chat-async") {
			t.Errorf("ai %s: help %q; want --timeout named exactly where it waits for the reply", use, cmd.Long)
		}
	}
	cmd, _, err := root.Find([]string{"ai", "agents", "export"})
	if err != nil || !strings.Contains(cmd.Short, "creates an export job") ||
		!strings.Contains(cmd.Long, "Emporix keeps each export as a job of type\nEXPORT") ||
		!strings.Contains(cmd.Long, "\n"+readScopedNote+"\n\n") ||
		!strings.HasSuffix(cmd.Long, "\n\nRequires the ai.agent_read scope.") {
		t.Errorf("ai agents export: %v, summary %q, help %q; want it to name its job and its read scope", err,
			cmd.Short, cmd.Long)
	}
}
