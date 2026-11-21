package cli

import (
	"bytes"
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/output"
)

func mixinSiteReads() []readCase {
	reads := siteAwareReads()
	cases := make([]readCase, len(reads))
	for i := range reads {
		cases[i] = reads[i].readCase()
	}
	return cases
}

const overlaySite = "AcmeStore_NL"

type siteAwareRead struct {
	siteAwareItem
	args    []string
	path    string
	scopes  []string
	method  string
	body    string
	paged   bool
	mode    readMode
	filters listFilters
}

func siteAwareReads() []siteAwareRead {
	products, instances := "/product/acme1/products", "/schema/acme1/custom-entities/CAR/instances"
	product := siteAwareRead{siteAwareItem: siteAwareProduct(), scopes: []string{productReadUnpublishedScope}}
	instance := siteAwareRead{siteAwareItem: siteAwareInstance(), scopes: []string{"custom.car_read"}}
	return []siteAwareRead{
		product.listed([]string{"products", "list"}, products),
		product.one([]string{"products", "get", "100200"}, products+"/100200"),
		product.searched([]string{"products", "search"}, products+"/search"),
		instance.listed([]string{"custom-entities", "instances", "list", "CAR"}, instances),
		instance.one([]string{"custom-entities", "instances", "get", "CAR", "CAR1"}, instances+"/CAR1"),
		instance.searched([]string{"custom-entities", "instances", "search", "CAR"}, instances+"/search"),
	}
}

func (r *siteAwareRead) one(args []string, path string) siteAwareRead {
	read := *r
	read.args, read.path, read.method, read.mode = args, path, http.MethodGet, getWhole
	return read
}

func (r *siteAwareRead) listed(args []string, path string) siteAwareRead {
	read := *r
	read.args, read.path, read.method, read.paged = args, path, http.MethodGet, true
	read.filters = listFilters{kinds: filterable}
	return read
}

func (r *siteAwareRead) searched(args []string, path string) siteAwareRead {
	read := *r
	read.args, read.path, read.method, read.body, read.paged = args, path, http.MethodPost, `{"q":""}`, true
	read.filters = listFilters{kinds: sortable}
	return read
}

func (r *siteAwareRead) sent(t *testing.T, extra http.Header) string {
	t.Helper()
	if r.paged {
		return pagedAs(t, r.method, r.path, r.body, extra)
	}
	return sentAs(t, r.method, r.path, r.body, extra)
}

func (r *siteAwareRead) readCase() readCase {
	return readCase{
		args: slices.Concat(r.args, []string{"--" + mixinSiteFlag, overlaySite}), path: r.path, scopes: r.scopes,
		mode: r.mode, filters: r.filters, method: r.method, sent: r.body, site: overlaySite,
		body: r.response(r.merged), want: r.tables[string(defaultLanguage)],
	}
}

func (r *siteAwareRead) response(item string) string {
	if r.paged {
		return "[" + item + "]"
	}
	return item
}

func (r *siteAwareRead) outputs() map[string][]string {
	all := map[string][]string{"table": nil, "csv": {"-o", "csv"}, "json": {"-o", "json"}, "yaml": {"-o", "yaml"}}
	if r.paged {
		all["stream"] = []string{"--" + streamFlag}
	}
	return all
}

func (r *siteAwareRead) printed(t *testing.T, merged bool, lang, format string) string {
	t.Helper()
	item := r.item(merged)
	switch format {
	case "table":
		return r.tables[lang]
	case "csv":
		return r.csv[lang]
	case "yaml":
		return r.yaml(merged, r.paged)
	case "stream":
		return item + "\n"
	case "json":
		var indented bytes.Buffer
		if err := json.Indent(&indented, []byte(r.response(item)), "", "  "); err != nil {
			t.Fatal(err)
		}
		return indented.String() + "\n"
	default:
		t.Fatalf("no output %q", format)
		return ""
	}
}

func (r *siteAwareRead) serve(fake *fakeEmporix) {
	unmerged, merged := r.response(r.item(false)), r.response(r.item(true))
	fake.route(r.path, func(w http.ResponseWriter, req *http.Request) {
		answer := unmerged
		switch site, sent := req.Header["Emporix-Site"]; {
		case sent && slices.Equal(site, []string{overlaySite}):
			answer = merged
		case sent:
			respond(http.StatusBadRequest, `{"message":"no overlays for this site"}`)(w, req)
			return
		}
		if r.paged {
			firstPage(answer)(w, req)
			return
		}
		respond(http.StatusOK, answer)(w, req)
	})
}

func epxHeader(extra http.Header) http.Header {
	header := http.Header{"Accept": {"application/json"}, "Accept-Encoding": {"gzip"}, "Accept-Language": {"*"},
		"Authorization": {"Bearer " + testToken}, "User-Agent": {"emporix-cli"}}
	maps.Copy(header, extra)
	return header
}

func sentAs(t *testing.T, method, uri, body string, extra http.Header) string {
	t.Helper()
	header := epxHeader(extra)
	if body != "" {
		header.Set("Content-Length", strconv.Itoa(len(body)))
		header.Set("Content-Type", "application/json")
	}
	return rendered(t, method, uri, header, body)
}

func pagedAs(t *testing.T, method, path, body string, extra http.Header) string {
	t.Helper()
	paged := http.Header{"X-Total-Count": {"true"}}
	maps.Copy(paged, extra)
	return sentAs(t, method, path+"?pageNumber=1&pageSize=100", body, paged) +
		sentAs(t, method, path+"?pageNumber=2&pageSize=100", body, paged)
}

func wire(t *testing.T, fake *fakeEmporix) string {
	t.Helper()
	var b strings.Builder
	for i, uri := range fake.gets {
		header := fake.headers[i].Clone()
		agent, _, _ := strings.Cut(header.Get("User-Agent"), "/")
		header.Set("User-Agent", agent)
		b.WriteString(rendered(t, fake.methods[i], uri, header, fake.bodies[i]))
	}
	return b.String()
}

func rendered(t *testing.T, method, uri string, header http.Header, body string) string {
	t.Helper()
	var sent []byte
	if body != "" {
		sent = []byte(body)
	}
	var b strings.Builder
	if err := output.WriteRequest(&b, method, uri, header, sent); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

type keySource struct {
	name   string
	vars   func(fake *fakeEmporix) map[string]string
	global []string
	lang   string
}

func keySources() []keySource {
	prod := func(*fakeEmporix) map[string]string { return map[string]string{envProfile: "prod"} }
	return []keySource{
		{name: "the environment's key", vars: (*fakeEmporix).env, lang: "en"},
		{name: "the default profile", lang: "nl"},
		{name: "EPX_PROFILE", vars: prod, lang: "en"},
		{name: "--profile", global: []string{"--" + profileFlag, "prod"}, lang: "en"},
	}
}

func runSiteAware(t *testing.T, read *siteAwareRead, source *keySource, args ...string) (
	fake *fakeEmporix, code int, stdout, stderr string) {
	t.Helper()
	fake = newFakeEmporix(t)
	read.serve(fake)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret, "prod": testSecret})
	var vars map[string]string
	if source.vars != nil {
		vars = source.vars(fake)
	}
	code, stdout, stderr = runIn(t, setup.env(vars, "", false), slices.Concat(source.global, read.args, args)...)
	return fake, code, stdout, stderr
}

func TestSiteAwareReadsWithoutTheFlagStayAsBefore(t *testing.T) {
	checkSiteAwareReads(t, "")
}

func TestSiteAwareReadsSendTheMixinSiteGiven(t *testing.T) {
	checkSiteAwareReads(t, overlaySite)
}

func checkSiteAwareReads(t *testing.T, site string) {
	t.Helper()
	var args []string
	var extra http.Header
	if site != "" {
		args, extra = []string{"--" + mixinSiteFlag, site}, http.Header{"Emporix-Site": {site}}
	}
	reads := siteAwareReads()
	for r := range reads {
		read := &reads[r]
		sent := read.sent(t, extra)
		for _, source := range keySources() {
			for format, flags := range read.outputs() {
				t.Run(strings.Join(read.args, " ")+"/"+source.name+"/"+format, func(t *testing.T) {
					fake, code, stdout, stderr := runSiteAware(t, read, &source, slices.Concat(args, flags)...)
					if want := read.printed(t, site != "", source.lang, format); code != ExitOK || stdout != want {
						t.Fatalf("exit %d, stderr %q, stdout:\n%s\nwant:\n%s", code, stderr, stdout, want)
					}
					if got := wire(t, fake); got != sent {
						t.Fatalf("sent:\n%s\nwant:\n%s", got, sent)
					}
				})
			}
		}
	}
}

func TestMixinSiteOnlyOnTheSiteAwareReads(t *testing.T) {
	root := newTestRoot()
	var want []string
	for _, read := range siteAwareReads() {
		cmd, _, err := root.Find(read.args)
		if err != nil {
			t.Fatalf("%q names no command: %v", read.args, err)
		}
		want = append(want, cmd.CommandPath())
		if !strings.Contains(cmd.Long, mixinSiteHelp) {
			t.Errorf("%q: help %q; want what --%s does", read.args, cmd.Long, mixinSiteFlag)
		}
	}
	slices.Sort(want)
	if got := commandsWithFlag(mixinSiteFlag); !slices.Equal(got, want) {
		t.Fatalf("--%s on %q; want %q", mixinSiteFlag, got, want)
	}
}

func TestMixinSiteRefusesWhatNamesNoSite(t *testing.T) {
	for _, value := range []string{"", " ", "\t", " main", "main\t", "main\r\nX-Injected: 1", "main\nX",
		"main\x7f"} {
		fake := newFakeEmporix(t)
		code, stdout, stderr := runCLI(t, fake.env(), "products", "list", "--"+mixinSiteFlag, value)
		if code != ExitUsage || stdout != "" || !strings.Contains(stderr, `for "--`+mixinSiteFlag+`" flag`) {
			t.Errorf("%q: exit %d, stdout %q, stderr %q; want a usage error of the flag", value, code, stdout, stderr)
		}
		fake.assertNothingSent(t)
	}
}
