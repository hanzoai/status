package api

import (
	"io/fs"
	"os"

	"github.com/TwiN/health"
	"github.com/TwiN/logr"
	"github.com/hanzoai/status/config"
	"github.com/hanzoai/status/config/ui"
	"github.com/hanzoai/status/config/web"
	static "github.com/hanzoai/status/web"
	"github.com/hanzoai/status/zipx"
	metric "github.com/luxfi/metric"
	fiber "github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/fiber/v3/middleware/compress"
	"github.com/zap-proto/fiber/v3/middleware/cors"
	"github.com/zap-proto/fiber/v3/middleware/recover"
	"github.com/zap-proto/fiber/v3/middleware/redirect"
	fiberstatic "github.com/zap-proto/fiber/v3/middleware/static"
	"github.com/zap-proto/zip"
)

type API struct {
	router *zip.App
}

func New(cfg *config.Config) *API {
	api := &API{}
	if cfg.Web == nil {
		logr.Warnf("[api.New] nil web config passed as parameter. This should only happen in tests. Using default web configuration")
		cfg.Web = web.GetDefaultConfig()
	}
	if cfg.UI == nil {
		logr.Warnf("[api.New] nil ui config passed as parameter. This should only happen in tests. Using default ui configuration")
		cfg.UI = ui.GetDefaultConfig()
	}
	api.router = api.createRouter(cfg)
	return api
}

func (a *API) Router() *zip.App {
	return a.router
}

func (a *API) createRouter(cfg *config.Config) *zip.App {
	app := zip.New(zip.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			logr.Errorf("[api.ErrorHandler] %s", err.Error())
			return fiber.DefaultErrorHandler(c, err)
		},
		ReadBufferSize: cfg.Web.ReadBufferSize,
	})
	if os.Getenv("ENVIRONMENT") == "dev" {
		app.Use(zipx.Wrap(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:8081"},
			AllowCredentials: true,
		})))
	}
	// Middlewares
	app.Use(zipx.Wrap(recover.New()))
	app.Use(zipx.Wrap(compress.New()))
	// Define metrics handler, if necessary
	if cfg.Metrics {
		metricsHandler := metric.InstrumentMetricHandler(metric.DefaultRegisterer, metric.NewHTTPHandler(metric.DefaultGatherer, metric.HandlerOpts{
			DisableCompression: true,
		}))
		app.Get("/metrics", zip.AdaptNetHTTP(metricsHandler))
	}
	// Define main router
	apiRouter := app.Group("/v1/status")
	////////////////////////
	// UNPROTECTED ROUTES //
	////////////////////////
	unprotectedAPIRouter := apiRouter.Group("/")
	unprotectedAPIRouter.Get("/config", ConfigHandler{securityConfig: cfg.Security, config: cfg}.GetConfig)
	unprotectedAPIRouter.Get("/endpoints/:key/health/badge.svg", HealthBadge)
	unprotectedAPIRouter.Get("/endpoints/:key/health/badge.shields", HealthBadgeShields)
	unprotectedAPIRouter.Get("/endpoints/:key/uptimes/:duration", UptimeRaw)
	unprotectedAPIRouter.Get("/endpoints/:key/uptimes/:duration/badge.svg", UptimeBadge)
	unprotectedAPIRouter.Get("/endpoints/:key/response-times/:duration", ResponseTimeRaw)
	unprotectedAPIRouter.Get("/endpoints/:key/response-times/:duration/badge.svg", ResponseTimeBadge(cfg))
	unprotectedAPIRouter.Get("/endpoints/:key/response-times/:duration/chart.svg", ResponseTimeChart)
	unprotectedAPIRouter.Get("/endpoints/:key/response-times/:duration/history", ResponseTimeHistory)
	// This endpoint requires authz with bearer token, so technically it is protected
	unprotectedAPIRouter.Post("/endpoints/:key/external", CreateExternalEndpointResult(cfg))
	// SPA
	app.Get("/", SinglePageApplication(cfg.UI))
	app.Get("/endpoints/:key", SinglePageApplication(cfg.UI))
	app.Get("/suites/:key", SinglePageApplication(cfg.UI))
	// The brand's marks, and the manifest that names them. Registered before the
	// static middleware so they answer the bare root paths a browser guesses at.
	RegisterBrandIcons(app, cfg.UI)
	app.Get("/manifest.json", Manifest(cfg.UI))
	// Health endpoint
	healthHandler := health.Handler().WithJSON(true)
	app.Get("/health", func(c *zip.Ctx) error {
		statusCode, body := healthHandler.GetResponseStatusCodeAndBody()
		return c.Bytes(statusCode, body)
	})
	// Custom CSS
	app.Get("/css/custom.css", CustomCSSHandler{customCSS: cfg.UI.CustomCSS}.GetCustomCSS)
	// Everything else falls back on static content
	app.Use(zipx.Wrap(redirect.New(redirect.Config{
		Rules: map[string]string{
			"/index.html": "/",
		},
		StatusCode: 301,
	})))
	staticFileSystem, err := fs.Sub(static.FileSystem, static.RootPath)
	if err != nil {
		panic(err)
	}
	app.Use(zip.H(func(c *zip.Ctx) error {
		// Static assets have no filename hashing, so ensure browsers revalidate
		path := c.Path()
		if len(path) > 3 {
			switch path[len(path)-3:] {
			case ".js":
				c.SetHeader("Cache-Control", "no-cache")
			}
		}
		if len(path) > 4 && path[len(path)-4:] == ".css" {
			c.SetHeader("Cache-Control", "no-cache")
		}
		return c.Next()
	}))
	app.Use(zipx.Wrap(fiberstatic.New("", fiberstatic.Config{
		FS:         staticFileSystem,
		IndexNames: []string{"index.html"},
		Browse:     true,
	})))
	//////////////////////
	// PROTECTED ROUTES //
	//////////////////////
	// ORDER IS IMPORTANT: all routes applied AFTER the security middleware will require authn
	protectedAPIRouter := apiRouter.Group("/")
	if cfg.Security != nil {
		if err := cfg.Security.RegisterHandlers(app); err != nil {
			panic(err)
		}
		if err := cfg.Security.ApplySecurityMiddleware(protectedAPIRouter); err != nil {
			panic(err)
		}
	}
	protectedAPIRouter.Get("/endpoints/statuses", EndpointStatuses(cfg))
	protectedAPIRouter.Get("/endpoints/:key/statuses", EndpointStatus(cfg))
	protectedAPIRouter.Get("/suites/statuses", SuiteStatuses(cfg))
	protectedAPIRouter.Get("/suites/:key/statuses", SuiteStatus(cfg))
	return app
}
