package controller

import (
	"os"
	"time"

	"github.com/TwiN/logr"
	"github.com/hanzoai/status/api"
	"github.com/hanzoai/status/config"
	fiber "github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/zip"
)

var (
	app *zip.App
)

// Handle creates the router and starts the server
func Handle(cfg *config.Config) {
	api := api.New(cfg)
	app = api.Router()
	server := app.Fiber().Server()
	server.ReadTimeout = 15 * time.Second
	server.WriteTimeout = 15 * time.Second
	server.IdleTimeout = 15 * time.Second
	if os.Getenv("ROUTER_TEST") == "true" {
		return
	}
	logr.Info("[controller.Handle] Listening on " + cfg.Web.SocketAddress())
	// TLS is a value in the listen config, not a second Listen method. zip's own
	// Listen carries neither a TLS nor a dual-stack knob, so this is the one
	// place the app escapes to the underlying fiber listener.
	listenConfig := fiber.ListenConfig{ListenerNetwork: fiber.NetworkTCP}
	if cfg.Web.HasTLS() {
		listenConfig.CertFile = cfg.Web.TLS.CertificateFile
		listenConfig.CertKeyFile = cfg.Web.TLS.PrivateKeyFile
	}
	if err := app.Fiber().Listen(cfg.Web.SocketAddress(), listenConfig); err != nil {
		logr.Fatalf("[controller.Handle] %s", err.Error())
	}
	logr.Info("[controller.Handle] Server has shut down successfully")
}

// Shutdown stops the server
func Shutdown() {
	if app != nil {
		_ = app.Fiber().Shutdown()
		app = nil
	}
}
