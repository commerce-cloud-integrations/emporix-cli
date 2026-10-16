package emporix

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestErrorMessage checks the message that errorMessage takes from each
// error shape: the message member, followed by the details in each shape
// that the Emporix references document, field and message only; it leaves
// out what a detail holds besides, a detail without a message, one that
// repeats the message and any other JSON; and it bounds the result.
func TestErrorMessage(t *testing.T) {
	const violation = `"message":"The request violates validation constraints.","type":"validation_violation"`
	checkErrorMessages(t, []errorMessageCase{
		{name: "service message", body: `{"message":"not found"}`, want: "not found"},
		{name: "oauth2 description", body: `{"error":"invalid_scope","error_description":"bad scope"}`, want: "bad scope"},
		{name: "gateway fault", body: `{"fault":{"faultstring":"Invalid Access Token"}}`, want: "Invalid Access Token"},
		{name: "unknown JSON object is omitted", body: `{"access_token":"SENTINEL"}`, want: ""},
		{name: "JSON array is omitted", body: `[{"token":"SENTINEL"}]`, want: ""},
		{name: "message text is kept for the caller to sanitize", body: `{"message":" a\u001bb "}`, want: "a\u001bb"},
		{name: "plain text kept", body: "  upstream timeout  ", want: "upstream timeout"},
		{name: "objects with a field (most services)",
			body: `{` + violation + `,"details":[{"field":"name","type":"invalid_field","message":"Cannot specify ` +
				`multiple localizations when 'Content-Language' header is set.","moreInfo":"SENTINEL"}]}`,
			want: "The request violates validation constraints. (name: Cannot specify multiple localizations when " +
				"'Content-Language' header is set.)"},
		{name: "several details", body: `{` + violation + `,"details":[{"field":"id","message":"must not be null"},` +
			`{"field":"fees","message":"may not be empty"}]}`,
			want: "The request violates validation constraints. (id: must not be null; fees: may not be empty)"},
		{name: "objects without a field (Returns, Search, Site Settings)",
			body: `{"message":"Bad request","details":[{"message":"unknown index"}]}`,
			want: "Bad request (unknown index)"},
		{name: "strings (Price, IAM, Quote and others)",
			body: `{"code":400,"message":"Invalid input","details":["siteCode is required"," currency is unknown "]}`,
			want: "Invalid input (siteCode is required; currency is unknown)"},
		{name: "one object (Customer Segment Service)",
			body: `{"message":"Invalid segment","details":{"field":"name","message":"must not be blank"}}`,
			want: "Invalid segment (name: must not be blank)"},
		{name: "details without a message", body: `{"code":400,"details":["unknown site"]}`, want: "unknown site"},
		{name: "other members of a detail are left out",
			body: `{"message":"m","details":[{"field":"f","type":"t","moreInfo":"SENTINEL","token":"SENTINEL"}]}`,
			want: "m"},
		{name: "a detail that repeats the message", body: `{"message":"m","details":["m"]}`, want: "m"},
		{name: "details of other types", body: `{"message":"m","details":[1,null,{"message":2},[{"message":"x"}]]}`,
			want: "m"},
		{name: "a detail kept for the caller to sanitize",
			body: `{"message":"m","details":[{"field":"a\u001bb","message":"c\u202ed"}]}`,
			want: "m (a\u001bb: c\u202ed)"},
		{name: "details of another shape", body: `{"message":"m","details":"SENTINEL"}`, want: "m"},
		{name: "a message trimmed before details repeat it", body: `{"message":" m ","details":["m","x"]}`,
			want: "m (x)"},
		{name: "a blank message leaves the next", body: `{"message":" ","error_description":"bad scope"}`,
			want: "bad scope"},
	})
	long := strings.Repeat("é", maxMessageBytes)
	detail := `{"field":"name","message":"` + strings.Repeat("é", 40) + `"}`
	for _, tt := range []struct{ body, start string }{
		{body: long, start: "é"},
		{body: `{"message":"` + long + `"}`, start: "é"},
		{body: `{"message":"invalid","details":[` + strings.Repeat(detail+",", 20) + detail + `]}`,
			start: "invalid (name: é"},
	} {
		got := errorMessage([]byte(tt.body))
		prefix, truncated := strings.CutSuffix(got, "…")
		if !truncated || len(prefix) > maxMessageBytes || !strings.HasPrefix(prefix, tt.start) || !utf8.ValidString(prefix) {
			t.Errorf("errorMessage = %q (%d bytes); want it bounded to %d bytes", got, len(got), maxMessageBytes)
		}
	}
}

// errorMessageCase is a body and the message that errorMessage takes from it.
type errorMessageCase struct{ name, body, want string }

func checkErrorMessages(t *testing.T, cases []errorMessageCase) {
	t.Helper()
	for _, tt := range cases {
		if got := errorMessage([]byte(tt.body)); got != tt.want {
			t.Errorf("%s: errorMessage = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestErrorMessageMembers checks which member gives the message: those of
// the services and OAuth before the gateway's faultstring, then
// problem+json's detail (a string only) and its title; no other member.
func TestErrorMessageMembers(t *testing.T) {
	checkErrorMessages(t, []errorMessageCase{
		{name: "problem detail (Reward Points Service)",
			body: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"Failed to read request",` +
				`"instance":"/reward-points/acme1/redeemOptions"}`,
			want: "Failed to read request"},
		{name: "problem title where it has no detail",
			body: `{"type":"about:blank","title":"Conflict","status":409,"detail":" ","instance":"SENTINEL"}`,
			want: "Conflict"},
		{name: "a detail that is no string", body: `{"title":"Bad Request","detail":{"token":"SENTINEL"}}`,
			want: "Bad Request"},
		{name: "gateway fault with its detail object",
			body: `{"fault":{"faultstring":"Invalid access token","detail":{"errorcode":"oauth.v2.InvalidAccessToken"}}}`,
			want: "Invalid access token"},
		{name: "a message before a detail", body: `{"message":"m","detail":"d"}`, want: "m"},
		{name: "OAuth members before a detail",
			body: `{"error":"invalid_request","error_description":"d1","detail":"d2","title":"t"}`, want: "d1"},
		{name: "a detail before a title", body: `{"title":"t","detail":"d"}`, want: "d"},
		{name: "no other member", body: `{"status":400,"path":"SENTINEL","type":"SENTINEL","instance":"SENTINEL"}`,
			want: ""},
	})
}

// TestErrorMessageErrorCode checks that a string errorCode follows the
// message in parentheses, unless the message names it as a word, and stands
// first without a message.
func TestErrorMessageErrorCode(t *testing.T) {
	checkErrorMessages(t, []errorMessageCase{
		{name: "after the message (Shipping Service)",
			body: `{"code":400,"status":"Bad Request","message":"Invalid delivery time","errorCode":"DELIVERY-TIME-400003"}`,
			want: "Invalid delivery time (DELIVERY-TIME-400003)"},
		{name: "before the details",
			body: `{"status":400,"message":"The request violates validation constraints.","type":"validation_violation",` +
				`"errorCode":"ENTITY-400001","details":[{"field":"shipTo[0].country","message":"not a valid country code",` +
				`"type":"invalid_field"}]}`,
			want: "The request violates validation constraints. (ENTITY-400001) (shipTo[0].country: not a valid country code)"},
		{name: "without a message", body: `{"code":400,"errorCode":"ENTITY-400001"}`, want: "ENTITY-400001"},
		{name: "with details, without a message", body: `{"errorCode":"E-1","details":["x"]}`, want: "E-1 (x)"},
		{name: "named by the message", body: `{"message":"m (E-1)","errorCode":"E-1"}`, want: "m (E-1)"},
		{name: "named by the message after a part of a longer word",
			body: `{"message":"E-10 or E-1","errorCode":"E-1"}`, want: "E-10 or E-1"},
		{name: "a part of a longer word of the message", body: `{"message":"m (E-10, XE-1)","errorCode":"E-1"}`,
			want: "m (E-10, XE-1) (E-1)"},
		{name: "named by the message right after a longer word", body: `{"message":"E-1E-1","errorCode":"E-1"}`,
			want: "E-1E-1 (E-1)"},
		{name: "null", body: `{"message":"m","errorCode":null,"id":"SENTINEL"}`, want: "m"},
		{name: "no string", body: `{"message":"m","errorCode":400}`, want: "m"},
	})
}

// TestErrorMessageDetailMembers checks the details of Micronaut's
// _embedded.errors and of fileDetails, read as those of a details member,
// after them and nothing else of _embedded.
func TestErrorMessageDetailMembers(t *testing.T) {
	checkErrorMessages(t, []errorMessageCase{
		{name: "embedded errors (Brand, Price, Schema, Unit Service)",
			body: `{"message":"Bad Request","_links":{"self":{"href":"SENTINEL","templated":false}},` +
				`"_embedded":{"errors":[{"message":"Failed to convert argument [name]","path":"/name",` +
				`"_links":{"self":{"href":"SENTINEL"}}},{"message":"Required Body [unitCodes] not specified"}]}}`,
			want: "Bad Request (/name: Failed to convert argument [name]; Required Body [unitCodes] not specified)"},
		{name: "an embedded error that repeats the message",
			body: `{"message":"Invalid JSON","_embedded":{"errors":[{"message":"Invalid JSON"},{"message":"x"}]}}`,
			want: "Invalid JSON (x)"},
		{name: "embedded errors without a message", body: `{"_embedded":{"errors":[{"message":"x","path":"/p"}]}}`,
			want: "/p: x"},
		{name: "embedded errors as strings", body: `{"message":"m","_embedded":{"errors":["x"]}}`, want: "m (x)"},
		{name: "other embedded members",
			body: `{"message":"m","_embedded":{"items":[{"message":"SENTINEL"}],"errors":"SENTINEL"}}`, want: "m"},
		{name: "embedded of another shape", body: `{"message":"m","_embedded":["SENTINEL"]}`, want: "m"},
		{name: "file details (Schema Service)",
			body: `{"code":400,"message":"Error while parsing file on tenant='acme1'","fileDetails":[{"fileName":"SENTINEL",` +
				`"message":"Unexpected character at line 3"}]}`,
			want: "Error while parsing file on tenant='acme1' (Unexpected character at line 3)"},
		{name: "each member, in order",
			body: `{"message":"m","fileDetails":{"message":"c"},"_embedded":{"errors":[{"message":"b"}]},"details":["a"]}`,
			want: "m (a; b; c)"},
	})
}

// TestErrorMessageDetailFields checks the field of a detail: its field or
// else its path, left out where the message starts with it as a word.
func TestErrorMessageDetailFields(t *testing.T) {
	checkErrorMessages(t, []errorMessageCase{
		{name: "a field before a path",
			body: `{"message":"m","_embedded":{"errors":[{"field":"f","path":"/p","message":"x"}]}}`, want: "m (f: x)"},
		{name: "a path of a detail", body: `{"message":"m","details":[{"path":"/p","message":"x"}]}`, want: "m (/p: x)"},
		{name: "a message that starts with its field (Availability Service)",
			body: `{"message":"m","details":[{"field":"productId","message":"productId: must not be blank"}]}`,
			want: "m (productId: must not be blank)"},
		{name: "a message that starts with its field and a space (Shipping Service)",
			body: `{"message":"m","details":[{"field":"shipTo[0].country","message":"shipTo[0].country Provided value ` +
				`is not valid"}]}`,
			want: "m (shipTo[0].country Provided value is not valid)"},
		{name: "a message that starts with a longer word than its field",
			body: `{"message":"m","details":[{"field":"name","message":"names must differ"}]}`,
			want: "m (name: names must differ)"},
		{name: "a message that is its field", body: `{"message":"m","details":[{"field":"id","message":"id"}]}`,
			want: "m (id)"},
	})
}
