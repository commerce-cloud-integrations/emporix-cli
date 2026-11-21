package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
)

const (
	testSecret = "top-secret-value"
	testToken  = "issued-access-token-sentinel"
)

type fakeEmporix struct {
	mu          sync.Mutex
	denied      []string
	routes      map[string]http.HandlerFunc
	gets        []string
	methods     []string
	totals      []string
	languages   []string
	sent        []string
	bodies      []string
	types       []string
	headers     []http.Header
	tokenScopes []string
	server      *httptest.Server
}

func newFakeEmporix(t *testing.T, denied ...string) *fakeEmporix {
	t.Helper()
	f := &fakeEmporix{denied: denied, routes: map[string]http.HandlerFunc{}}
	f.server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeEmporix) route(path string, handler http.HandlerFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routes[path] = handler
}

func (f *fakeEmporix) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.URL.Path == "/oauth/token" {
		f.token(w, r)
		return
	}
	f.gets = append(f.gets, r.URL.RequestURI())
	f.methods = append(f.methods, r.Method)
	f.totals = append(f.totals, r.Header.Get("X-Total-Count"))
	f.languages = append(f.languages, headerValue(r.Header, "Content-Language"))
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respond(http.StatusBadRequest, `{"message":"unreadable body"}`)(w, r)
		return
	}
	f.sent = append(f.sent, strings.TrimSpace(r.Method+" "+string(body)))
	f.bodies = append(f.bodies, string(body))
	f.types = append(f.types, r.Header.Get("Content-Type"))
	f.headers = append(f.headers, r.Header.Clone())
	r.Body = io.NopCloser(strings.NewReader(string(body)))
	if handler, ok := f.routes[r.URL.EscapedPath()]; ok {
		handler(w, r)
		return
	}
	respond(http.StatusNotFound, `{"message":"not found"}`)(w, r)
}

func (f *fakeEmporix) token(w http.ResponseWriter, r *http.Request) {
	f.tokenScopes = append(f.tokenScopes, r.PostFormValue("scope"))
	scopes := strings.Fields(r.PostFormValue("scope"))
	switch {
	case r.PostFormValue("client_secret") != testSecret:
		respond(http.StatusUnauthorized, `{"Error":"Client credentials are invalid"}`)(w, r)
	case len(scopes) == 0 || scopes[0] != "tenant=acme1":
		respond(http.StatusBadRequest, `{"message":"The tenant in the scope does not match"}`)(w, r)
	default:
		granted := slices.DeleteFunc(scopes, func(s string) bool { return slices.Contains(f.denied, s) })
		respond(http.StatusOK, `{"access_token":"`+testToken+`","token_type":"Bearer","expires_in":3600,"scope":"`+
			strings.Join(granted, " ")+`"}`)(w, r)
	}
}

const emptyHeader = "(empty)"

func headerValue(h http.Header, name string) string {
	if values, sent := h[http.CanonicalHeaderKey(name)]; sent && strings.Join(values, "") == "" {
		return emptyHeader
	}
	return h.Get(name)
}

func respond(status int, body string, headers ...string) http.HandlerFunc {
	if len(headers)%2 != 0 {
		panic("respond: header names and values must pair")
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		for i := 0; i < len(headers); i += 2 {
			w.Header().Set(headers[i], headers[i+1])
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func inTurn(handlers ...http.HandlerFunc) http.HandlerFunc {
	sent := 0
	return func(w http.ResponseWriter, r *http.Request) {
		handlers[min(sent, len(handlers)-1)](w, r)
		sent++
	}
}

func (tt *readCase) served(t *testing.T) http.HandlerFunc {
	t.Helper()
	switch {
	case tt.contentType != "":
		return inLanguage(tt.language(), respond(http.StatusOK, tt.response, "Content-Type", tt.contentType))
	case tt.totalHeader != "":
		return allLanguages(respond(http.StatusOK, "", "X-Total-Count", tt.totalHeader))
	case tt.mode == getAllBodyPaged:
		return allLanguages(firstBodyPage(tt.response))
	case tt.mode == getAllOnFirstPage:
		var items []json.RawMessage
		if err := json.Unmarshal([]byte(tt.response), &items); err != nil {
			t.Fatalf("a getAllOnFirstPage case's body must be an array: %v", err)
		}
		return allLanguages(respond(http.StatusOK, tt.response, "X-Total-Count", strconv.Itoa(len(items))))
	default:
		return allLanguages(firstPage(tt.response))
	}
}

func (tt *readCase) language() string {
	if i := slices.Index(tt.args, "--"+languageFlag); i >= 0 && i+1 < len(tt.args) {
		return tt.args[i+1]
	}
	return ""
}

func firstPage(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := body
		if number := r.URL.Query().Get("pageNumber"); number != "" && number != "1" {
			page = "[]"
		}
		respond(http.StatusOK, page)(w, r)
	}
}

func firstBodyPage(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var paging struct {
			PageNumber int `json:"pageNumber"`
		}
		page := "[]"
		if json.NewDecoder(r.Body).Decode(&paging) == nil && paging.PageNumber == 1 {
			page = body
		}
		respond(http.StatusOK, page)(w, r)
	}
}

func allLanguages(handler http.HandlerFunc) http.HandlerFunc {
	return inLanguage("*", handler)
}

func inLanguage(language string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Join(r.Header.Values("Accept-Language"), ",") != language {
			respond(http.StatusOK, `[{"id":"001234","name":"single language only"}]`)(w, r)
			return
		}
		handler(w, r)
	}
}

func withoutPaging(t *testing.T, uri string) (path string, query url.Values) {
	t.Helper()
	sent, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	query = sent.Query()
	query.Del("pageNumber")
	query.Del("pageSize")
	return sent.EscapedPath(), query
}

func (f *fakeEmporix) assertScopes(t *testing.T, scopes ...string) {
	t.Helper()
	want := strings.Join(append([]string{"tenant=acme1"}, scopes...), " ")
	if len(f.tokenScopes) == 0 || slices.ContainsFunc(f.tokenScopes, func(s string) bool { return s != want }) {
		t.Fatalf("token scopes %q; want %q", f.tokenScopes, want)
	}
}

func (f *fakeEmporix) assertHeader(t *testing.T, name, want string) {
	t.Helper()
	for i, header := range f.headers {
		if sent := headerValue(header, name); sent != want {
			t.Fatalf("request %d sent %s %q; want %q", i+1, name, sent, want)
		}
	}
}

func (f *fakeEmporix) assertSent(t *testing.T, method, body string) {
	t.Helper()
	for i, sent := range f.sent {
		if want := strings.TrimSpace(method + " " + strings.ReplaceAll(body, "{page}", strconv.Itoa(i+1))); sent != want {
			t.Fatalf("request %d sent %q; want %q", i+1, sent, want)
		}
	}
	if len(f.sent) == 0 {
		t.Fatal("no request sent")
	}
}

type readMode int

const (
	getAll readMode = iota
	getWhole
	getAllWithoutTotal
	getAllInOnePage
	getAllOnFirstPage
	getAllPageObjects
	getAllHybrisCounted
	getAllBodyPaged
	getAllContentPages
)

func (f *fakeEmporix) assertMode(t *testing.T, mode readMode, page, path, query string) {
	t.Helper()
	if (mode == getAllInOnePage || mode == getAllOnFirstPage) != (page != "") {
		t.Fatalf("a one-page case, and only one, names its page size (page %q)", page)
	}
	want, err := url.ParseQuery(query)
	if err != nil || want.Encode() != query {
		t.Fatalf("query %q; write it encoded, with sorted keys", query)
	}
	uri := path
	if query != "" {
		uri += "?" + query
	}
	switch mode {
	case getWhole:
		f.assertRequests(t, uri)
		return
	case getAllBodyPaged:
		f.assertRequests(t, uri, uri)
		if !slices.Equal(f.totals, []string{"true", "true"}) {
			t.Fatalf("X-Total-Count %q; want it asked on both pages", f.totals)
		}
		return
	case getAllPageObjects:
		f.assertFirstPageObject(t, path, want, "1")
		return
	case getAllContentPages:
		f.assertFirstPageObject(t, path, want, "0")
		return
	case getAllHybrisCounted:
		f.assertPagesOf(t, "100", "", "")
		want.Set("totalCount", "true")
	case getAllWithoutTotal:
		f.assertPagesOf(t, "100", "", "")
	case getAllInOnePage:
		f.assertPagesOf(t, page, "true", "true")
	case getAllOnFirstPage:
		f.assertPagesOf(t, page, "true")
	default:
		f.assertPagesOf(t, "100", "true", "true")
	}
	assertQueries(t, f.gets, path, want)
}

func (f *fakeEmporix) assertFirstPageObject(t *testing.T, path string, query url.Values, first string) {
	t.Helper()
	query.Set("page", first)
	query.Set("size", "100")
	f.assertRequests(t, path+"?"+query.Encode())
	if !slices.Equal(f.totals, []string{""}) {
		t.Fatalf("X-Total-Count %q; want none", f.totals)
	}
}

func (f *fakeEmporix) assertPagesOf(t *testing.T, size string, totals ...string) {
	t.Helper()
	pages := make([]string, len(totals))
	for i := range pages {
		pages[i] = strconv.Itoa(i + 1)
	}
	assertPages(t, f.gets, pages...)
	if !slices.Equal(f.totals, totals) {
		t.Fatalf("X-Total-Count %q; want %q", f.totals, totals)
	}
	for _, get := range f.gets {
		if sent, err := url.Parse(get); err != nil || sent.Query().Get("pageSize") != size {
			t.Fatalf("request %q; want pages of %s items", get, size)
		}
	}
}

func (f *fakeEmporix) assertRequests(t *testing.T, uris ...string) {
	t.Helper()
	if !slices.Equal(f.gets, uris) {
		t.Fatalf("requests %q; want only %q", f.gets, uris)
	}
}

func assertQueries(t *testing.T, gets []string, path string, query url.Values) {
	t.Helper()
	if len(gets) == 0 {
		t.Fatal("no request")
	}
	for _, get := range gets {
		if sentPath, sentQuery := withoutPaging(t, get); sentPath != path || sentQuery.Encode() != query.Encode() {
			t.Fatalf("request %q; want %s with %q", get, path, query.Encode())
		}
	}
}

func assertPages(t *testing.T, gets []string, pages ...string) {
	t.Helper()
	sent := make([]string, len(gets))
	for i, get := range gets {
		if parsed, err := url.Parse(get); err == nil {
			sent[i] = parsed.Query().Get("pageNumber")
		}
	}
	if !slices.Equal(sent, pages) {
		t.Fatalf("pages %q; want %q", sent, pages)
	}
}

type startedReader struct {
	once    sync.Once
	started chan struct{}
	r       io.Reader
}

func (s *startedReader) Read(p []byte) (int, error) {
	s.once.Do(func() { close(s.started) })
	return s.r.Read(p)
}

func runUntilReading(t *testing.T, args []string, env map[string]string, terminal bool) (<-chan int, *bytes.Buffer) {
	t.Helper()
	pending, writer := io.Pipe()
	t.Cleanup(func() { _ = writer.Close() })
	stdin := &startedReader{started: make(chan struct{}), r: pending}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	errOut := &bytes.Buffer{}
	exited := make(chan int, 1)
	go func() {
		exited <- Run(ctx, args, Env{
			In: stdin, Out: io.Discard, Err: errOut, Getenv: func(key string) string { return env[key] },
			Terminal: terminal,
		})
	}()
	select {
	case <-stdin.started:
		cancel()
	case code := <-exited:
		t.Fatalf("exit %d before reading standard input, stderr %q", code, errOut.String())
	}
	return exited, errOut
}

func (f *fakeEmporix) env() map[string]string {
	return map[string]string{
		envAPIURL: f.server.URL, envTenant: "acme1", envClientID: "client", envClientSecret: testSecret,
	}
}

func assertNoSecret(t *testing.T, outputs ...string) {
	t.Helper()
	for _, text := range outputs {
		if strings.Contains(text, testSecret) || strings.Contains(text, testToken) {
			t.Fatalf("output leaks credential material: %q", text)
		}
	}
}

func TestFakeTellsAnEmptyHeaderFromNone(t *testing.T) {
	for _, tt := range []struct {
		header http.Header
		want   string
	}{
		{header: http.Header{}, want: ""},
		{header: http.Header{"Content-Language": {""}}, want: emptyHeader},
		{header: http.Header{"Content-Language": {"nl"}}, want: "nl"},
	} {
		if got := headerValue(tt.header, "Content-Language"); got != tt.want {
			t.Errorf("headerValue(%v) = %q; want %q", tt.header, got, tt.want)
		}
	}
}
