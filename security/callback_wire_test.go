package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/zap-proto/zip"
	"golang.org/x/oauth2"
)

// The OIDC callback is the one place a browser turns an authorization code into
// a session cookie, so its answer is a security surface and not just an answer:
// the status a browser reads, the exact refusal it is told, and — on the one
// path that succeeds — the cookie it is handed. This table pins all of that
// BYTE FOR BYTE, so a change of transport underneath is provably not a change
// of behaviour on the wire.
//
// Every case is driven through the real router (app.Fiber().Test), not by
// calling the handler directly, because the transport is exactly what is under
// test. The identity provider is local and deterministic: an RSA key signs the
// id_token, oidc.StaticKeySet verifies it, and a httptest server plays the
// token endpoint. Nothing here reaches the network.

// wire is one expected answer. A field left at its zero value is not asserted,
// EXCEPT status and body, which every case states.
type wire struct {
	status      int
	contentType string
	body        string
	// nosniff records whether X-Content-Type-Options: nosniff is expected.
	nosniff  bool
	location string
	// cookie is the name of a Set-Cookie expected on the response.
	cookie string
}

func (w wire) check(t *testing.T, resp *http.Response, body string) {
	t.Helper()
	if resp.StatusCode != w.status {
		t.Errorf("status = %d, want %d", resp.StatusCode, w.status)
	}
	if got := resp.Header.Get("Content-Type"); got != w.contentType {
		t.Errorf("Content-Type = %q, want %q", got, w.contentType)
	}
	if body != w.body {
		t.Errorf("body = %q, want %q", body, w.body)
	}
	want := ""
	if w.nosniff {
		want = "nosniff"
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != want {
		t.Errorf("X-Content-Type-Options = %q, want %q", got, want)
	}
	if got := resp.Header.Get("Location"); got != w.location {
		t.Errorf("Location = %q, want %q", got, w.location)
	}
	name := ""
	for _, c := range resp.Cookies() {
		if c.Name == w.cookie {
			name = c.Name
		}
	}
	if w.cookie != "" && name == "" {
		t.Errorf("no Set-Cookie named %q; got %v", w.cookie, resp.Cookies())
	}
	if w.cookie == "" && len(resp.Cookies()) != 0 {
		t.Errorf("unexpected Set-Cookie: %v", resp.Cookies())
	}
}

// idp is a local identity provider: it signs id_tokens and answers the token
// endpoint, so the success path can be exercised without a network.
type idp struct {
	key      *rsa.PrivateKey
	issuer   string
	clientID string
	server   *httptest.Server
	// token is what the token endpoint returns; a test sets it per case.
	token string
	// noIDToken omits the id_token field entirely, which is a different thing
	// from sending it empty: an empty string still type-asserts to a string, so
	// the handler carries on to the verifier instead of refusing early. Measured.
	noIDToken bool
}

func newIDP(t *testing.T) *idp {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	p := &idp{key: key, issuer: "https://idp.test", clientID: "client-id"}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body := map[string]any{"access_token": "access", "token_type": "Bearer"}
		if !p.noIDToken {
			body["id_token"] = p.token
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(p.server.Close)
	return p
}

// sign builds an RS256 JWT over claims. Hand-rolled because a JWT is three
// base64 segments and a signature, and the standard library signs.
func (p *idp) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	seg := func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(b)
	}
	signing := seg(map[string]any{"alg": "RS256", "typ": "JWT"}) + "." + seg(claims)
	sum := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// config returns an OIDCConfig wired to this provider, bypassing initialize()
// (which performs discovery over the network) by setting the two fields it
// would have set.
func (p *idp) config(allowed ...string) *OIDCConfig {
	return &OIDCConfig{
		IssuerURL:       p.issuer,
		RedirectURL:     "http://localhost:80/authorization-code/callback",
		ClientID:        p.clientID,
		ClientSecret:    "client-secret",
		Scopes:          []string{"openid"},
		AllowedSubjects: allowed,
		SessionTTL:      DefaultOIDCSessionTTL,
		oauth2Config: oauth2.Config{
			ClientID:     p.clientID,
			ClientSecret: "client-secret",
			Scopes:       []string{"openid"},
			RedirectURL:  "http://localhost:80/authorization-code/callback",
			Endpoint:     oauth2.Endpoint{TokenURL: p.server.URL + "/token", AuthURL: p.server.URL + "/auth"},
		},
		verifier: oidc.NewVerifier(p.issuer,
			&oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&p.key.PublicKey}},
			&oidc.Config{ClientID: p.clientID}),
	}
}

// callbackApp registers the callback on a router exactly as RegisterHandlers
// does. This one line is the transport under test; the table below never moves.
func callbackApp(c *OIDCConfig) *zip.App {
	app := zip.New(zip.Config{})
	app.All("/authorization-code/callback", zip.AdaptNetHTTP(http.HandlerFunc(c.callbackHandler)))
	return app
}

func ask(t *testing.T, app *zip.App, method, target string, cookies ...*http.Cookie) (*http.Response, string) {
	t.Helper()
	req := httptest.NewRequest(method, target, http.NoBody)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := app.Fiber().Test(req)
	if err != nil {
		t.Fatalf("Test(%s %s): %v", method, target, err)
	}
	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, string(b)
}

const callback = "/authorization-code/callback"

func TestCallbackWire_Refusals(t *testing.T) {
	p := newIDP(t)
	app := callbackApp(p.config())

	for _, tc := range []struct {
		name    string
		method  string
		target  string
		cookies []*http.Cookie
		want    wire
	}{
		{
			name:   "provider reported an error",
			method: "GET",
			target: callback + "?error=access_denied&error_description=user+said+no",
			want: wire{
				status:      http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "access_denied: user said no\n",
				nosniff:     true,
			},
		},
		{
			name:   "no state cookie",
			method: "GET",
			target: callback,
			want: wire{
				status:      http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "state not found\n",
				nosniff:     true,
			},
		},
		{
			name:    "state did not match",
			method:  "GET",
			target:  callback + "?state=attacker",
			cookies: []*http.Cookie{{Name: cookieNameState, Value: "browser"}},
			want: wire{
				status:      http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "state did not match\n",
				nosniff:     true,
			},
		},
		{
			// The route is All, so form_post callbacks arrive as POST. The
			// refusal must read the same to a browser either way.
			name:    "state did not match, POST",
			method:  "POST",
			target:  callback + "?state=attacker",
			cookies: []*http.Cookie{{Name: cookieNameState, Value: "browser"}},
			want: wire{
				status:      http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "state did not match\n",
				nosniff:     true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := ask(t, app, tc.method, tc.target, tc.cookies...)
			tc.want.check(t, resp, body)
		})
	}
}

// State and nonce are compared against a cookie, and an empty value compares
// equal to an empty value. So a cookie the browser holds as empty — which any
// sibling host can plant, HttpOnly being no defence against writing one — turns
// both comparisons into no-ops: the state check that exists to stop a login
// CSRF, and the nonce that binds the id_token to this browser's own login.
// Neither may be satisfiable by having nothing.
func TestCallbackWire_EmptyStateAndNonce(t *testing.T) {
	p := newIDP(t)

	t.Run("an empty state is not a matching state", func(t *testing.T) {
		resp, body := ask(t, callbackApp(p.config()), "GET", callback+"?state=",
			&http.Cookie{Name: cookieNameState, Value: ""})
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want 400 — an empty state must not reach the token exchange", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
			t.Errorf("Content-Type = %q, want text/plain; charset=utf-8", got)
		}
		if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
		}
		if body != "state not found\n" {
			t.Errorf("body = %q, want %q", body, "state not found\n")
		}
		if len(resp.Cookies()) != 0 {
			t.Errorf("unexpected Set-Cookie: %v", resp.Cookies())
		}
	})

	t.Run("an absent nonce claim is not a matching nonce", func(t *testing.T) {
		c := p.config()
		// A well-formed, correctly signed id_token that simply carries no nonce.
		p.token = p.sign(t, map[string]any{
			"iss": p.issuer, "aud": p.clientID, "sub": "anyone@example.com",
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
		})
		resp, body := ask(t, callbackApp(c), "GET", callback+"?state=s&code=c",
			&http.Cookie{Name: cookieNameState, Value: "s"},
			&http.Cookie{Name: cookieNameNonce, Value: ""})
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want 400 — an unbound id_token must not mint a session", resp.StatusCode)
		}
		if body != "nonce not found\n" {
			t.Errorf("body = %q, want %q", body, "nonce not found\n")
		}
		for _, ck := range resp.Cookies() {
			if ck.Name == cookieNameSession {
				t.Error("a session was handed out for an id_token bound to no login")
			}
		}
	})
}

func TestCallbackWire_TokenNotUsable(t *testing.T) {
	p := newIDP(t)
	c := p.config()
	app := callbackApp(c)
	state := []*http.Cookie{{Name: cookieNameState, Value: "s"}}

	t.Run("id_token missing from the token response", func(t *testing.T) {
		p.noIDToken = true
		defer func() { p.noIDToken = false }()
		resp, body := ask(t, app, "GET", callback+"?state=s&code=c", state...)
		wire{
			status:      http.StatusInternalServerError,
			contentType: "text/plain; charset=utf-8",
			body:        "Missing 'id_token' in oauth2 token\n",
			nosniff:     true,
		}.check(t, resp, body)
	})

	t.Run("id_token does not verify", func(t *testing.T) {
		p.token = "not.a.jwt"
		resp, body := ask(t, app, "GET", callback+"?state=s&code=c", state...)
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
			t.Errorf("Content-Type = %q, want text/plain; charset=utf-8", got)
		}
		if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
		}
		// The verifier's message is its own; what this pins is the prefix the
		// handler puts in front of it, and that the answer ends in a newline.
		if !strings.HasPrefix(body, "Failed to verify id_token: ") || !strings.HasSuffix(body, "\n") {
			t.Errorf("body = %q, want %q-prefixed and newline-terminated", body, "Failed to verify id_token: ")
		}
		if len(resp.Cookies()) != 0 {
			t.Errorf("unexpected Set-Cookie on a token that did not verify: %v", resp.Cookies())
		}
	})
}

func TestCallbackWire_Nonce(t *testing.T) {
	p := newIDP(t)
	c := p.config()
	app := callbackApp(c)
	p.token = p.sign(t, map[string]any{
		"iss":   p.issuer,
		"aud":   p.clientID,
		"sub":   "user1@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"nonce": "from-the-provider",
	})

	t.Run("no nonce cookie", func(t *testing.T) {
		resp, body := ask(t, app, "GET", callback+"?state=s&code=c",
			&http.Cookie{Name: cookieNameState, Value: "s"})
		wire{
			status:      http.StatusBadRequest,
			contentType: "text/plain; charset=utf-8",
			body:        "nonce not found\n",
			nosniff:     true,
		}.check(t, resp, body)
	})

	t.Run("nonce did not match", func(t *testing.T) {
		resp, body := ask(t, app, "GET", callback+"?state=s&code=c",
			&http.Cookie{Name: cookieNameState, Value: "s"},
			&http.Cookie{Name: cookieNameNonce, Value: "a-different-one"})
		wire{
			status:      http.StatusBadRequest,
			contentType: "text/plain; charset=utf-8",
			body:        "nonce did not match\n",
			nosniff:     true,
		}.check(t, resp, body)
	})
}

// signedIn is the shape of the two answers that hand out a session, and the one
// that refuses to. These are the cases where the wire carries a credential.
func TestCallbackWire_SignedIn(t *testing.T) {
	p := newIDP(t)
	nonce := &http.Cookie{Name: cookieNameNonce, Value: "n"}
	state := &http.Cookie{Name: cookieNameState, Value: "s"}
	claims := func(sub string) map[string]any {
		return map[string]any{
			"iss": p.issuer, "aud": p.clientID, "sub": sub,
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
			"nonce": "n",
		}
	}

	t.Run("no allowed subjects means every subject is allowed", func(t *testing.T) {
		c := p.config()
		p.token = p.sign(t, claims("anyone@example.com"))
		resp, body := ask(t, callbackApp(c), "GET", callback+"?state=s&code=c", state, nonce)
		wire{
			status:      http.StatusFound,
			contentType: "text/html; charset=utf-8",
			body:        "<a href=\"/\">Found</a>.\n\n",
			location:    "/",
			cookie:      cookieNameSession,
		}.check(t, resp, body)
	})

	t.Run("an allowed subject", func(t *testing.T) {
		c := p.config("USER1@example.com") // the comparison is case-insensitive
		p.token = p.sign(t, claims("user1@example.com"))
		resp, body := ask(t, callbackApp(c), "GET", callback+"?state=s&code=c", state, nonce)
		wire{
			status:      http.StatusFound,
			contentType: "text/html; charset=utf-8",
			body:        "<a href=\"/\">Found</a>.\n\n",
			location:    "/",
			cookie:      cookieNameSession,
		}.check(t, resp, body)
	})

	t.Run("a subject that is not allowed gets no session", func(t *testing.T) {
		c := p.config("user1@example.com")
		p.token = p.sign(t, claims("intruder@example.com"))
		resp, body := ask(t, callbackApp(c), "GET", callback+"?state=s&code=c", state, nonce)
		wire{
			status:      http.StatusFound,
			contentType: "text/html; charset=utf-8",
			body:        "<a href=\"/?error=access_denied\">Found</a>.\n\n",
			location:    "/?error=access_denied",
		}.check(t, resp, body)
	})

	// The route is All, so form_post callbacks arrive as POST. The redirect and
	// the cookie are the same; the courtesy body is not written for a POST, and
	// the content type is the transport's default for a bodyless answer.
	t.Run("POST is redirected without the courtesy body", func(t *testing.T) {
		c := p.config()
		p.token = p.sign(t, claims("anyone@example.com"))
		resp, body := ask(t, callbackApp(c), "POST", callback+"?state=s&code=c", state, nonce)
		wire{
			status:      http.StatusFound,
			contentType: "text/plain; charset=utf-8",
			body:        "",
			location:    "/",
			cookie:      cookieNameSession,
		}.check(t, resp, body)
	})
}

// The session cookie is the credential this whole exchange exists to hand out.
// Its attributes are what keep it from being sent by a foreign site or read by
// a script, so they are pinned separately from the status line.
func TestCallbackWire_SessionCookieAttributes(t *testing.T) {
	p := newIDP(t)
	c := p.config()
	p.token = p.sign(t, map[string]any{
		"iss": p.issuer, "aud": p.clientID, "sub": "anyone@example.com",
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": "n",
	})
	resp, _ := ask(t, callbackApp(c), "GET", callback+"?state=s&code=c",
		&http.Cookie{Name: cookieNameState, Value: "s"},
		&http.Cookie{Name: cookieNameNonce, Value: "n"})

	var session *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == cookieNameSession {
			session = ck
		}
	}
	if session == nil {
		t.Fatalf("no %s cookie; got %v", cookieNameSession, resp.Cookies())
	}
	if session.Path != "/" {
		t.Errorf("Path = %q, want %q", session.Path, "/")
	}
	if session.MaxAge != int(DefaultOIDCSessionTTL.Seconds()) {
		t.Errorf("MaxAge = %d, want %d", session.MaxAge, int(DefaultOIDCSessionTTL.Seconds()))
	}
	if session.SameSite != http.SameSiteStrictMode {
		t.Errorf("SameSite = %v, want Strict", session.SameSite)
	}
	// The value is a session id the server can resolve, not the subject.
	if session.Value == "anyone@example.com" {
		t.Error("the cookie carries the subject itself")
	}
	if subject, ok := sessions.Get(session.Value); !ok || subject != "anyone@example.com" {
		t.Errorf("sessions[%q] = %v, %v; want the subject", session.Value, subject, ok)
	}
}
