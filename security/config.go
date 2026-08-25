package security

import (
	"encoding/base64"
	"net/http"

	fiber "github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/fiber/v3/middleware/basicauth"
	"github.com/zap-proto/zip"
	"golang.org/x/crypto/bcrypt"
	"hanzo.ai/status/zipx"
)

const (
	cookieNameState   = "gatus_state"
	cookieNameNonce   = "gatus_nonce"
	cookieNameSession = "gatus_session"

	// unauthorized is the body this surface has always answered a request
	// without a session with. Pinned in protect_wire_test.go.
	unauthorized = "token is missing or invalid"
)

// Router is the slice of zip's routing surface this package needs. Both
// *zip.App and zip.Router (a Group) satisfy it, so RegisterHandlers can take
// the app while ApplySecurityMiddleware takes a sub-router.
//
// Use takes Components — middleware, or another App — because that is the one
// composition verb zip has; the route methods still take Handlers.
type Router interface {
	Use(cs ...zip.Component) zip.Router
	All(path string, handlers ...zip.Handler) zip.Router
}

// Config is the security configuration for Gatus
type Config struct {
	Basic *BasicConfig `yaml:"basic,omitempty"`
	OIDC  *OIDCConfig  `yaml:"oidc,omitempty"`

	// session records that this config protects the surface with OIDC
	// sessions, set when the middleware is applied. Basic auth carries no
	// session, so IsAuthenticated answers for it in the negative.
	session bool
}

// ValidateAndSetDefaults returns whether the security configuration is valid or not and sets default values.
func (c *Config) ValidateAndSetDefaults() bool {
	return (c.Basic != nil && c.Basic.isValid()) || (c.OIDC != nil && c.OIDC.ValidateAndSetDefaults())
}

// RegisterHandlers registers all handlers required based on the security configuration
func (c *Config) RegisterHandlers(router Router) error {
	if c.OIDC != nil {
		if err := c.OIDC.initialize(); err != nil {
			return err
		}
		router.All("/oidc/login", c.OIDC.loginHandler)
		router.All("/authorization-code/callback", c.OIDC.callbackHandler)
	}
	return nil
}

// ApplySecurityMiddleware applies an authentication middleware to the router passed.
// The router passed should be a sub-router in charge of handlers that require authentication.
func (c *Config) ApplySecurityMiddleware(router Router) error {
	if c.OIDC != nil {
		c.session = true
		router.Use(zip.Handler(func(ctx *zip.Ctx) error {
			if !live(ctx) {
				return ctx.String(http.StatusUnauthorized, unauthorized)
			}
			return ctx.Next()
		}))
	} else if c.Basic != nil {
		var decodedBcryptHash []byte
		if len(c.Basic.PasswordBcryptHashBase64Encoded) > 0 {
			var err error
			decodedBcryptHash, err = base64.URLEncoding.DecodeString(c.Basic.PasswordBcryptHashBase64Encoded)
			if err != nil {
				return err
			}
		}
		router.Use(zipx.Wrap(basicauth.New(basicauth.Config{
			Authorizer: func(username, password string, _ fiber.Ctx) bool {
				if len(c.Basic.PasswordBcryptHashBase64Encoded) > 0 {
					if username != c.Basic.Username || bcrypt.CompareHashAndPassword(decodedBcryptHash, []byte(password)) != nil {
						return false
					}
				}
				return true
			},
			Unauthorized: func(ctx fiber.Ctx) error {
				ctx.Set("WWW-Authenticate", "Basic")
				return ctx.Status(401).SendString("Unauthorized")
			},
		})))
	}
	return nil
}

// live reports whether the request carries a session the server still holds.
// It is the ONE predicate: the middleware admits on it and IsAuthenticated
// reports it, so what a browser is told about itself can never disagree with
// what it is actually allowed to read.
func live(ctx *zip.Ctx) bool {
	_, held := sessions.Get(ctx.Fiber().Cookies(cookieNameSession))
	return held
}

// IsAuthenticated checks whether the user is authenticated
// If the Config does not warrant authentication, it will always return true.
func (c *Config) IsAuthenticated(ctx *zip.Ctx) bool {
	return c.session && live(ctx)
}
