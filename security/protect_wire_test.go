package security

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zap-proto/zip"
)

// ApplySecurityMiddleware is what stands between the internet and the routes
// registered after it. Three things about it are load-bearing, and only one of
// them was pinned before: the refusal (401 and its body), the ADMISSION of a
// live session — a middleware that refused everything would have satisfied the
// old test — and the barrier property, that routes registered BEFORE it stay
// public. All three are asserted here on exact bytes.

func protectedApp(t *testing.T, c *Config) *zip.App {
	t.Helper()
	app := zip.New(zip.Config{})
	// Registered before the middleware: the unprotected half of the surface.
	app.Get("/public", func(c *zip.Ctx) error { return c.String(http.StatusOK, "public") })
	if err := c.ApplySecurityMiddleware(app); err != nil {
		t.Fatalf("ApplySecurityMiddleware: %v", err)
	}
	app.Get("/private", func(c *zip.Ctx) error { return c.String(http.StatusOK, "private") })
	return app
}

func oidcSessionConfig() *Config {
	return &Config{OIDC: &OIDCConfig{
		IssuerURL:    "https://idp.test",
		RedirectURL:  "http://localhost:80/authorization-code/callback",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scopes:       []string{"openid"},
		SessionTTL:   DefaultOIDCSessionTTL,
	}}
}

func TestProtectWire(t *testing.T) {
	c := oidcSessionConfig()
	app := protectedApp(t, c)

	// A session the server actually knows, established the way the callback
	// establishes one.
	sessions.SetWithTTL("live-session", "user1@example.com", DefaultOIDCSessionTTL)
	t.Cleanup(func() { sessions.Delete("live-session") })

	for _, tc := range []struct {
		name    string
		path    string
		cookies []*http.Cookie
		status  int
		body    string
	}{
		{
			name:   "no cookie at all",
			path:   "/private",
			status: http.StatusUnauthorized,
			body:   "token is missing or invalid",
		},
		{
			name:    "a session id the server does not know",
			path:    "/private",
			cookies: []*http.Cookie{{Name: cookieNameSession, Value: "invented"}},
			status:  http.StatusUnauthorized,
			body:    "token is missing or invalid",
		},
		{
			name:    "the wrong cookie name carries no session",
			path:    "/private",
			cookies: []*http.Cookie{{Name: "session", Value: "live-session"}},
			status:  http.StatusUnauthorized,
			body:    "token is missing or invalid",
		},
		{
			// A bearer is not how this surface is entered; only the cookie is.
			name:    "a live session is admitted",
			path:    "/private",
			cookies: []*http.Cookie{{Name: cookieNameSession, Value: "live-session"}},
			status:  http.StatusOK,
			body:    "private",
		},
		{
			name:   "a route registered before the middleware stays public",
			path:   "/public",
			status: http.StatusOK,
			body:   "public",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, http.NoBody)
			for _, ck := range tc.cookies {
				req.AddCookie(ck)
			}
			resp, err := app.Fiber().Test(req)
			if err != nil {
				t.Fatalf("Test: %v", err)
			}
			b, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode != tc.status {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.status)
			}
			if string(b) != tc.body {
				t.Errorf("body = %q, want %q", string(b), tc.body)
			}
		})
	}
}

// The Authorization header is the default way g8 reads a token. This surface
// replaces that with the session cookie, so a bearer must buy nothing — pinned
// because a conversion that reverted to the default extractor would otherwise
// look like it worked.
func TestProtectWire_BearerBuysNothing(t *testing.T) {
	c := oidcSessionConfig()
	app := protectedApp(t, c)
	sessions.SetWithTTL("bearer-session", "user1@example.com", DefaultOIDCSessionTTL)
	t.Cleanup(func() { sessions.Delete("bearer-session") })

	req := httptest.NewRequest("GET", "/private", http.NoBody)
	req.Header.Set("Authorization", "Bearer bearer-session")
	resp, err := app.Fiber().Test(req)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 — a bearer is not a session here", resp.StatusCode)
	}
	if string(b) != "token is missing or invalid" {
		t.Errorf("body = %q, want %q", string(b), "token is missing or invalid")
	}
}
