package security

import (
	"net/http"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/zap-proto/zip"
)

func TestOIDCConfig_ValidateAndSetDefaults(t *testing.T) {
	c := &OIDCConfig{
		IssuerURL:       "https://sso.gatus.io/",
		RedirectURL:     "http://localhost:80/authorization-code/callback",
		ClientID:        "client-id",
		ClientSecret:    "client-secret",
		Scopes:          []string{"openid"},
		AllowedSubjects: []string{"user1@example.com"},
		SessionTTL:      0, // Not set! ValidateAndSetDefaults should set it to DefaultOIDCSessionTTL
	}
	if !c.ValidateAndSetDefaults() {
		t.Error("OIDCConfig should be valid")
	}
	if c.SessionTTL != DefaultOIDCSessionTTL {
		t.Error("expected SessionTTL to be set to DefaultOIDCSessionTTL")
	}
}

func TestOIDCConfig_callbackHandler(t *testing.T) {
	c := &OIDCConfig{
		IssuerURL:       "https://sso.gatus.io/",
		RedirectURL:     "http://localhost:80/authorization-code/callback",
		ClientID:        "client-id",
		ClientSecret:    "client-secret",
		Scopes:          []string{"openid"},
		AllowedSubjects: []string{"user1@example.com"},
	}
	if err := c.initialize(); err != nil {
		t.Fatal("expected no error, but got", err)
	}
	app := callbackApp(c)
	// Try with no state cookie
	resp, _ := ask(t, app, "GET", callback)
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("expected code to be 400, but was", resp.StatusCode)
	}
	// Try with state cookie
	resp, _ = ask(t, app, "GET", callback, &http.Cookie{Name: cookieNameState, Value: "fake-state"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("expected code to be 400, but was", resp.StatusCode)
	}
	// Try with state cookie and state query parameter
	resp, _ = ask(t, app, "GET", callback+"?state=fake-state", &http.Cookie{Name: cookieNameState, Value: "fake-state"})
	// Exchange should fail, so 500.
	if resp.StatusCode != http.StatusInternalServerError {
		t.Error("expected code to be 500, but was", resp.StatusCode)
	}
}

// sessionCookieOf drives setSessionCookie through a route and returns the
// cookie it wrote, which is the only way to observe what reaches a browser.
func sessionCookieOf(t *testing.T, c *OIDCConfig, subject string) *http.Cookie {
	t.Helper()
	app := zip.New(zip.Config{})
	app.Get("/session", func(ctx *zip.Ctx) error {
		c.setSessionCookie(ctx, &oidc.IDToken{Subject: subject})
		return ctx.NoContent(http.StatusOK)
	})
	resp, _ := ask(t, app, "GET", "/session")
	for _, ck := range resp.Cookies() {
		if ck.Name == cookieNameSession {
			return ck
		}
	}
	t.Fatalf("no %s cookie; got %v", cookieNameSession, resp.Cookies())
	return nil
}

func TestOIDCConfig_setSessionCookie(t *testing.T) {
	if got := sessionCookieOf(t, &OIDCConfig{}, "test@example.com"); got == nil {
		t.Error("expected cookie to be set")
	}
}

func TestOIDCConfig_setSessionCookieWithCustomTTL(t *testing.T) {
	customTTL := 30 * time.Minute
	sessionCookie := sessionCookieOf(t, &OIDCConfig{SessionTTL: customTTL}, "test@example.com")
	if sessionCookie.MaxAge != int(customTTL.Seconds()) {
		t.Errorf("expected cookie MaxAge to be %d, but was %d", int(customTTL.Seconds()), sessionCookie.MaxAge)
	}
}
