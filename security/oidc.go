package security

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/TwiN/logr"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	fiber "github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/zip"
	"golang.org/x/oauth2"
)

const (
	DefaultOIDCSessionTTL = 8 * time.Hour
)

// OIDCConfig is the configuration for OIDC authentication
type OIDCConfig struct {
	IssuerURL       string        `yaml:"issuer-url"`   // e.g. https://dev-12345678.okta.com
	RedirectURL     string        `yaml:"redirect-url"` // e.g. http://localhost:8080/authorization-code/callback
	ClientID        string        `yaml:"client-id"`
	ClientSecret    string        `yaml:"client-secret"`
	Scopes          []string      `yaml:"scopes"`           // e.g. ["openid"]
	AllowedSubjects []string      `yaml:"allowed-subjects"` // e.g. ["user1@example.com"]. If empty, all subjects are allowed
	SessionTTL      time.Duration `yaml:"session-ttl"`      // e.g. 8h. Defaults to 8 hours

	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

// ValidateAndSetDefaults returns whether the OIDC configuration is valid and sets default values.
func (c *OIDCConfig) ValidateAndSetDefaults() bool {
	if c.SessionTTL <= 0 {
		c.SessionTTL = DefaultOIDCSessionTTL
	}
	return len(c.IssuerURL) > 0 && len(c.RedirectURL) > 0 && strings.HasSuffix(c.RedirectURL, "/authorization-code/callback") && len(c.ClientID) > 0 && len(c.ClientSecret) > 0 && len(c.Scopes) > 0
}

func (c *OIDCConfig) initialize() error {
	provider, err := oidc.NewProvider(context.Background(), c.IssuerURL)
	if err != nil {
		return err
	}
	c.verifier = provider.Verifier(&oidc.Config{ClientID: c.ClientID})
	// Configure an OpenID Connect aware OAuth2 client.
	c.oauth2Config = oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Scopes:       c.Scopes,
		RedirectURL:  c.RedirectURL,
		Endpoint:     provider.Endpoint(),
	}
	return nil
}

func (c *OIDCConfig) loginHandler(ctx *zip.Ctx) error {
	state, nonce := uuid.NewString(), uuid.NewString()
	ctx.Fiber().Cookie(&fiber.Cookie{
		Name:     cookieNameState,
		Value:    state,
		Path:     "/",
		MaxAge:   int(time.Hour.Seconds()),
		SameSite: "lax",
		HTTPOnly: true,
	})
	ctx.Fiber().Cookie(&fiber.Cookie{
		Name:     cookieNameNonce,
		Value:    nonce,
		Path:     "/",
		MaxAge:   int(time.Hour.Seconds()),
		SameSite: "lax",
		HTTPOnly: true,
	})
	return ctx.Redirect(http.StatusFound, c.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce)))
}

// callbackHandler finishes the login. It answers three questions in order —
// did this browser start the login (state), was the id_token minted for that
// same login (nonce), and is this subject allowed here — and only then hands
// out a session.
//
// State and nonce are each compared against a cookie, so an EMPTY value must
// never satisfy the comparison: "" equals "", and a cookie planted empty would
// turn both checks into no-ops. Any host under the site's domain can write a
// cookie the browser will then send here, HttpOnly being no defence against
// writing one, so the length check is what keeps state a CSRF check and nonce
// a binding to this browser's own login rather than to any token of the right
// audience.
func (c *OIDCConfig) callbackHandler(ctx *zip.Ctx) error {
	if failure := ctx.Query("error"); len(failure) > 0 {
		return refuse(ctx, http.StatusBadRequest, failure+": "+ctx.Query("error_description"))
	}
	state := ctx.Fiber().Cookies(cookieNameState)
	if len(state) == 0 {
		return refuse(ctx, http.StatusBadRequest, "state not found")
	}
	if ctx.Query("state") != state {
		return refuse(ctx, http.StatusBadRequest, "state did not match")
	}
	oauth2Token, err := c.oauth2Config.Exchange(ctx.Context(), ctx.Query("code"))
	if err != nil {
		return refuse(ctx, http.StatusInternalServerError, "Error exchanging token: "+err.Error())
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return refuse(ctx, http.StatusInternalServerError, "Missing 'id_token' in oauth2 token")
	}
	idToken, err := c.verifier.Verify(ctx.Context(), rawIDToken)
	if err != nil {
		return refuse(ctx, http.StatusInternalServerError, "Failed to verify id_token: "+err.Error())
	}
	nonce := ctx.Fiber().Cookies(cookieNameNonce)
	if len(nonce) == 0 {
		return refuse(ctx, http.StatusBadRequest, "nonce not found")
	}
	if idToken.Nonce != nonce {
		return refuse(ctx, http.StatusBadRequest, "nonce did not match")
	}
	if !c.allows(idToken.Subject) {
		logr.Debugf("[security.callbackHandler] Subject %s is not in the list of allowed subjects", idToken.Subject)
		return ctx.Redirect(http.StatusFound, "/?error=access_denied")
	}
	c.setSessionCookie(ctx, idToken)
	return ctx.Redirect(http.StatusFound, "/")
}

// allows reports whether subject may sign in. An empty list allows everyone.
func (c *OIDCConfig) allows(subject string) bool {
	if len(c.AllowedSubjects) == 0 {
		return true
	}
	for _, allowed := range c.AllowedSubjects {
		if strings.EqualFold(allowed, subject) {
			return true
		}
	}
	return false
}

// refuse answers with the reason on its own line, typed as text and marked so
// that no browser sniffs something executable out of a message that quotes
// values the caller supplied.
func refuse(ctx *zip.Ctx, code int, reason string) error {
	ctx.SetHeader(fiber.HeaderContentType, "text/plain; charset=utf-8")
	ctx.SetHeader(fiber.HeaderXContentTypeOptions, "nosniff")
	return ctx.String(code, reason+"\n")
}

func (c *OIDCConfig) setSessionCookie(ctx *zip.Ctx, idToken *oidc.IDToken) {
	// At this point, the user has been confirmed. All that's left to do is create a session.
	sessionID := uuid.NewString()
	sessions.SetWithTTL(sessionID, idToken.Subject, c.SessionTTL)
	ctx.Fiber().Cookie(&fiber.Cookie{
		Name:     cookieNameSession,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(c.SessionTTL.Seconds()),
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}
