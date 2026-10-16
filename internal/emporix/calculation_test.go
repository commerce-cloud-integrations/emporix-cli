package emporix

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// TestOnlyAcceptCreatedTakesCreated checks that 201 Created answers only a
// read of a client made by AcceptCreated, as Emporix answers unit conversions
// sent as PUT: any other read, with a body or not, fails on it, unclassified,
// naming the status, since it may have created something.
func TestOnlyAcceptCreatedTakesCreated(t *testing.T) {
	const converted = `{"output":{"targetUnitAmount":0.1,"targetUnit":"kg"}}`
	const path = "/unit-handling/{tenant}/units/convert-unit-commands"
	for _, tt := range []struct {
		name    string
		client  func(c *Client) *Client
		failure string // the error; empty when the read takes 201
	}{
		{name: "PUT accepting Created", client: func(c *Client) *Client {
			return withBody(t, c, http.MethodPut, `{"input":{}}`).AcceptCreated()
		}},
		{name: "PUT", client: func(c *Client) *Client { return withBody(t, c, http.MethodPut, `{"input":{}}`) },
			failure: "PUT /unit-handling/acme1/units/convert-unit-commands: 201 Created"},
		{name: "POST", client: func(c *Client) *Client { return withBody(t, c, http.MethodPost, `{}`) },
			failure: "POST /unit-handling/acme1/units/convert-unit-commands: 201 Created"},
		{name: "GET", client: func(c *Client) *Client { return c },
			failure: "GET /unit-handling/acme1/units/convert-unit-commands: 201 Created"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/unit-handling/acme1/units/convert-unit-commands", respond(http.StatusCreated, converted))
		read := tt.client(fake.client(t))
		body, err := read.Get(context.Background(), path, nil)
		if tt.failure == "" && (err != nil || string(body) != converted) {
			t.Errorf("%s: body %s, error %v; want the conversion", tt.name, body, err)
		}
		if tt.failure != "" && (body != nil || err == nil || errors.Is(err, ErrNotFound) || errors.Is(err, ErrAuth) ||
			err.Error() != tt.failure) {
			t.Errorf("%s: body %s, error %v; want an unclassified %q", tt.name, body, err, tt.failure)
		}
	}
}

// TestCheckReadsAnAnswerOfItsStatus checks Check, the read of a check that
// Emporix answers with its status alone: an empty success, 204 No Content
// included, is no body and no error, a body is returned as sent, and a
// failure is classified as Get classifies it; 201 Created is one, as for any
// read but AcceptCreated's.
func TestCheckReadsAnAnswerOfItsStatus(t *testing.T) {
	const path = "/coupon/{tenant}/coupons/SAVE7/validation"
	for _, tt := range []struct {
		status    int
		body      string
		want      string // the body returned
		class     error  // the error class, when the check fails
		errorText string
	}{
		{status: http.StatusOK, body: ""},
		{status: http.StatusOK, body: " \n"},
		{status: http.StatusNoContent, body: ""},
		{status: http.StatusOK, body: `{"valid":true}`, want: `{"valid":true}`},
		{status: http.StatusNotFound, body: `{"message":"Coupon not found"}`, class: ErrNotFound,
			errorText: "POST /coupon/acme1/coupons/SAVE7/validation: 404 Not Found: Coupon not found"},
		{status: http.StatusBadRequest, body: `{"message":"Coupon expired"}`,
			errorText: "POST /coupon/acme1/coupons/SAVE7/validation: 400 Bad Request: Coupon expired"},
		// What creating a redemption answers: never a check that passed.
		{status: http.StatusCreated, body: `{"id":"r1"}`,
			errorText: "POST /coupon/acme1/coupons/SAVE7/validation: 201 Created"},
	} {
		fake := newFakeEmporix(t)
		fake.route("/coupon/acme1/coupons/SAVE7/validation", respond(tt.status, tt.body))
		check := withBody(t, fake.client(t), http.MethodPost, `{"orderTotal":{"amount":10,"currency":"EUR"}}`)
		body, err := check.Check(context.Background(), path, nil)
		switch {
		case tt.errorText == "" && (err != nil || string(body) != tt.want || (tt.want == "" && body != nil)):
			t.Errorf("%d %q: body %q, error %v; want %q", tt.status, tt.body, body, err, tt.want)
		case tt.errorText != "" && (body != nil || err == nil || !strings.Contains(err.Error(), tt.errorText) ||
			(tt.class != nil && !errors.Is(err, tt.class))):
			t.Errorf("%d %q: body %q, error %v; want %q", tt.status, tt.body, body, err, tt.errorText)
		}
		if fake.requests[0].Method != http.MethodPost {
			t.Errorf("sent %s; want POST", fake.requests[0].Method)
		}
	}
}
