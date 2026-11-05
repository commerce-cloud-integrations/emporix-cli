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

// The reads of items with site-aware mixins, products and custom instances,
// send the emporix-site header only with --mixin-site: without it, every
// request and every output stays as it was before the flag, also where a
// profile gives a default site.

// mixinSiteReads returns the read cases of the six reads that take
// --mixin-site, each with --mixin-site overlaySite: the site whose overlays
// Emporix merges (emporix-site); their tables stay as without it.
func mixinSiteReads() []readCase {
	reads := siteAwareReads()
	cases := make([]readCase, len(reads))
	for i := range reads {
		cases[i] = reads[i].readCase()
	}
	return cases
}

// overlaySite is the site whose overlays the fake merges, as Emporix does
// for emporix-site: the default site of the profile dev too (twoProfiles).
const overlaySite = "AcmeStore_NL"

// siteAwareRead is a read of items with site-aware mixins, served by the
// fake at path: the item or items unmerged, or merged for overlaySite, as
// Emporix answers without and with emporix-site.
type siteAwareRead struct {
	siteAwareItem
	args    []string    // the read, without --mixin-site
	path    string      // the path that the fake serves, as escaped on the wire
	scopes  []string    // the scopes besides the tenant
	method  string      // the method of every request
	body    string      // the body of every request, e.g. a search's; none when empty
	paged   bool        // a collection: the item on page 1, none after
	mode    readMode    // how it reads (readCase.mode)
	filters listFilters // what it takes besides --limit (readCase.filters)
}

// siteAwareReads returns the reads that take --mixin-site, each of which
// every test of --mixin-site checks.
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

// one returns a copy of r, the read args of one item at path, a GET.
func (r *siteAwareRead) one(args []string, path string) siteAwareRead {
	read := *r
	read.args, read.path, read.method, read.mode = args, path, http.MethodGet, getWhole
	return read
}

// listed returns a copy of r, the read args of the items at path, every page
// read with a GET, filtered and sorted as a list.
func (r *siteAwareRead) listed(args []string, path string) siteAwareRead {
	read := *r
	read.args, read.path, read.method, read.paged = args, path, http.MethodGet, true
	read.filters = listFilters{kinds: filterable}
	return read
}

// searched returns a copy of r, the read args of the items that a search at
// path finds, every page read with a POST of an empty q filter, sorted as a
// search.
func (r *siteAwareRead) searched(args []string, path string) siteAwareRead {
	read := *r
	read.args, read.path, read.method, read.body, read.paged = args, path, http.MethodPost, `{"q":""}`, true
	read.filters = listFilters{kinds: sortable}
	return read
}

// sent returns every request of the read as wire renders it, with extra
// besides epx's own headers.
func (r *siteAwareRead) sent(t *testing.T, extra http.Header) string {
	t.Helper()
	if r.paged {
		return pagedAs(t, r.method, r.path, r.body, extra)
	}
	return sentAs(t, r.method, r.path, r.body, extra)
}

// readCase returns the read case of the read with --mixin-site overlaySite:
// the item merged, its table in the default language.
func (r *siteAwareRead) readCase() readCase {
	return readCase{
		args: slices.Concat(r.args, []string{"--" + mixinSiteFlag, overlaySite}), path: r.path, scopes: r.scopes,
		mode: r.mode, filters: r.filters, method: r.method, sent: r.body, site: overlaySite,
		body: r.response(r.merged), want: r.tables[string(defaultLanguage)],
	}
}

// response returns what the fake answers for the read: the item, or for a
// collection an array of it.
func (r *siteAwareRead) response(item string) string {
	if r.paged {
		return "[" + item + "]"
	}
	return item
}

// outputs returns the outputs of the read checked, by name: its table, -o
// csv, -o json and -o yaml, and, for a collection, --stream.
func (r *siteAwareRead) outputs() map[string][]string {
	all := map[string][]string{"table": nil, "csv": {"-o", "csv"}, "json": {"-o", "json"}, "yaml": {"-o", "yaml"}}
	if r.paged {
		all["stream"] = []string{"--" + streamFlag}
	}
	return all
}

// printed returns what the read prints in format, the item merged or not:
// the table or CSV in lang, the response indented by two spaces (json), as
// YAML, or the item as one line of JSON (stream); its keys sorted as they are.
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

// serve routes the read on fake: unmerged without emporix-site, merged for
// overlaySite, and 400 Bad Request for any other site, as no other has an
// overlay here, so that a site sent by mistake fails the read.
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

// epxHeader returns the headers that every read of epx sends, as wire
// renders them, User-Agent without its version, with extra besides.
func epxHeader(extra http.Header) http.Header {
	header := http.Header{"Accept": {"application/json"}, "Accept-Encoding": {"gzip"}, "Accept-Language": {"*"},
		"Authorization": {"Bearer " + testToken}, "User-Agent": {"emporix-cli"}}
	maps.Copy(header, extra)
	return header
}

// sentAs returns how wire renders one request of method to uri with body:
// with epx's headers, those of the body and extra besides.
func sentAs(t *testing.T, method, uri, body string, extra http.Header) string {
	t.Helper()
	header := epxHeader(extra)
	if body != "" {
		header.Set("Content-Length", strconv.Itoa(len(body)))
		header.Set("Content-Type", "application/json")
	}
	return rendered(t, method, uri, header, body)
}

// pagedAs returns how wire renders the requests of a read of every page of
// the collection at path, which the fake serves on page 1 (firstPage): pages
// 1 and 2 of 100, each asking for the total, with extra besides.
func pagedAs(t *testing.T, method, path, body string, extra http.Header) string {
	t.Helper()
	paged := http.Header{"X-Total-Count": {"true"}}
	maps.Copy(paged, extra)
	return sentAs(t, method, path+"?pageNumber=1&pageSize=100", body, paged) +
		sentAs(t, method, path+"?pageNumber=2&pageSize=100", body, paged)
}

// wire renders every request that fake received, in their order, as
// rendered renders it, User-Agent without its version.
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

// rendered renders one request as a dry run shows it (output.WriteRequest),
// without a body where body is empty.
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

// keySource is a source of the API key of a run that may give a default
// site: the environment's key, which gives none, the default profile dev
// (site AcmeStore_NL, the site of the overlays, and language nl) or prod
// (site main), which EPX_PROFILE or --profile names.
type keySource struct {
	name   string
	vars   func(fake *fakeEmporix) map[string]string // the environment of the run; none when nil
	global []string                                  // the global flags of the run
	lang   string                                    // the language of its tables
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

// runSiteAware serves read on a new fake and runs it with the key of source
// and args after the read's own; it returns the fake, the exit code and the
// output.
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

// TestSiteAwareReadsWithoutTheFlagStayAsBefore pins every request and the
// output, table, -o csv, -o json, -o yaml and --stream, of the six reads
// without --mixin-site, whatever gives the API key: no emporix-site header,
// the base mixins and their overlays printed as stored, also where the
// profile in effect has a default site.
func TestSiteAwareReadsWithoutTheFlagStayAsBefore(t *testing.T) {
	checkSiteAwareReads(t, "")
}

// TestSiteAwareReadsSendTheMixinSiteGiven checks that the six reads send the
// site that --mixin-site gives, not the profile's default site (main for
// prod), as emporix-site, and otherwise what they send without it, and
// print the mixins that Emporix merged for it, in the same table and CSV.
func TestSiteAwareReadsSendTheMixinSiteGiven(t *testing.T) {
	checkSiteAwareReads(t, overlaySite)
}

// checkSiteAwareReads runs each of the six reads with --mixin-site site
// (none where site is empty), with the key of every source, in each of its
// outputs, and checks that it prints the item, merged with the flag, as that
// output prints it, and sends exactly what it sends without the flag, with
// the header emporix-site site besides.
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

// TestMixinSiteOnlyOnTheSiteAwareReads checks that --mixin-site is on the
// site-aware reads (siteAwareReads) alone, never on a write, so that no read
// takes it without the tests of --mixin-site, and that their help says what
// the overlays are.
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

// TestMixinSiteRefusesWhatNamesNoSite checks that a value of --mixin-site
// that is blank, so names no site, has spaces around the site code, or
// cannot be sent as a header (a line break, DEL) fails as a usage error,
// named by its flag, before anything is sent.
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
