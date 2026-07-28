package endpoint

import "testing"

// spaShell is what api.hanzo.ai returns for any path the cloud binary does not
// route: the embedded console's static index, HTTP 200. Captured verbatim from
// https://api.hanzo.ai/health, which is the URL the board's own top-line
// "api.hanzo.ai" row used to probe.
const spaShell = `<!DOCTYPE html><!--v8.5.31--><html lang="en" class="t_dark" ` +
	`style="background-color:#000000;color-scheme:dark"><head><meta charSet="utf-8"/>` +
	`<title>Hanzo Cloud Console</title></head><body></body></html>`

// The whole /v1 surface can be down while the shell keeps answering 200 — and
// under a status-code-only condition the board reads UP. That is not a
// hypothetical: it is why the Hanzo board stayed green through a cloud outage.
//
// This pins the fix. A board row is only allowed to conclude "healthy" from a
// fact the shell cannot produce.
func TestStatusAloneCannotDistinguishTheSPAShellFromALiveAPI(t *testing.T) {
	shell := &Result{HTTPStatus: 200, Body: []byte(spaShell)}
	api := &Result{HTTPStatus: 200, Body: []byte(`{"status":"ok"}`)}

	statusOnly := Condition("[STATUS] == 200")
	if !statusOnly.evaluate(shell, false, false, nil) {
		t.Fatal("premise wrong: the shell does answer 200")
	}
	if !statusOnly.evaluate(api, false, false, nil) {
		t.Fatal("premise wrong: a live API answers 200")
	}
	// Both pass. The condition carries no information about which one it got.

	body := Condition("[BODY].status == ok")
	if body.evaluate(shell, false, false, nil) {
		t.Error("[BODY].status == ok accepted the console shell — the board would " +
			"still read UP with the API down")
	}
	if !body.evaluate(api, false, false, nil) {
		t.Error("[BODY].status == ok rejected a healthy API response")
	}
}

// The gated subsystems (sql, kv, vector, search, functions, o11y, …) answer
// 401/403 with a JSON envelope. Their rows accept the gate as UP, so the gate's
// SHAPE is the only thing separating "the subsystem is mounted and refusing me"
// from "the subsystem is gone and the shell replied".
func TestGatedSubsystemRowsRejectTheShell(t *testing.T) {
	shell := &Result{HTTPStatus: 200, Body: []byte(spaShell)}
	gated := &Result{HTTPStatus: 403, Body: []byte(`{"status":403,"error":"X-Org-Id required"}`)}

	for _, c := range []Condition{
		"has([BODY].status) == true",
		"has([BODY].error) == true",
	} {
		if c.evaluate(shell, false, false, nil) {
			t.Errorf("%s accepted the console shell", c)
		}
	}
	if !Condition("has([BODY].status) == true").evaluate(gated, false, false, nil) {
		t.Error("has([BODY].status) rejected a real tenancy gate")
	}
	if !Condition("has([BODY].error) == true").evaluate(gated, false, false, nil) {
		t.Error("has([BODY].error) rejected a real tenancy gate")
	}
}

// A catalog that comes up empty is the other silent failure: 200, valid JSON,
// correct shape, zero models. Customers see an empty picker; the board sees a
// healthy row.
func TestEmptyCatalogIsNotHealthy(t *testing.T) {
	empty := &Result{HTTPStatus: 200, Body: []byte(`{"object":"list","data":[]}`)}
	served := &Result{HTTPStatus: 200, Body: []byte(`{"object":"list","data":[{"id":"zen-embedding"}]}`)}

	c := Condition("len([BODY].data) > 0")
	if c.evaluate(empty, false, false, nil) {
		t.Error("len([BODY].data) > 0 accepted an empty model catalog")
	}
	if !c.evaluate(served, false, false, nil) {
		t.Error("len([BODY].data) > 0 rejected a served catalog")
	}
}

// IAM answering with the wrong issuer breaks every login while every endpoint
// stays up — a 200 tells you nothing, the issuer tells you everything.
func TestIAMIssuerIsPinned(t *testing.T) {
	right := &Result{HTTPStatus: 200, Body: []byte(`{"issuer":"https://hanzo.id"}`)}
	wrong := &Result{HTTPStatus: 200, Body: []byte(`{"issuer":"https://iam.hanzo.ai"}`)}

	c := Condition("[BODY].issuer == https://hanzo.id")
	if !c.evaluate(right, false, false, nil) {
		t.Error("pinned issuer rejected the canonical issuer")
	}
	if c.evaluate(wrong, false, false, nil) {
		t.Error("pinned issuer accepted the legacy issuer — logins would be broken and green")
	}
}
