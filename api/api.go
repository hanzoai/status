package api

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/TwiN/health"
	"github.com/TwiN/logr"
	metric "github.com/luxfi/metric"
	fiber "github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/fiber/v3/middleware/compress"
	"github.com/zap-proto/fiber/v3/middleware/cors"
	"github.com/zap-proto/fiber/v3/middleware/recover"
	"github.com/zap-proto/fiber/v3/middleware/redirect"
	fiberstatic "github.com/zap-proto/fiber/v3/middleware/static"
	"github.com/zap-proto/zip"
	"hanzo.ai/status/config"
	"hanzo.ai/status/config/ui"
	"hanzo.ai/status/config/web"
	static "hanzo.ai/status/web"
	"hanzo.ai/status/zipx"
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
		app.Raw(http.MethodGet, "/metrics", scrape(metric.DefaultRegisterer))
	}
	// Define main router
	apiRouter := app.Group("/v1/status")
	////////////////////////
	// UNPROTECTED ROUTES //
	////////////////////////
	unprotectedAPIRouter := apiRouter.Group("/")
	unprotectedAPIRouter.Raw(http.MethodGet, "/config", ConfigHandler{securityConfig: cfg.Security, config: cfg}.GetConfig)
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/health/badge.svg", HealthBadge)
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/health/badge.shields", HealthBadgeShields)
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/uptimes/:duration", UptimeRaw)
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/uptimes/:duration/badge.svg", UptimeBadge)
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/response-times/:duration", ResponseTimeRaw)
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/response-times/:duration/badge.svg", ResponseTimeBadge(cfg))
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/response-times/:duration/chart.svg", ResponseTimeChart)
	unprotectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/response-times/:duration/history", ResponseTimeHistory)
	// This endpoint requires authz with bearer token, so technically it is protected
	unprotectedAPIRouter.Raw(http.MethodPost, "/endpoints/:key/external", CreateExternalEndpointResult(cfg))
	// SPA
	app.Raw(http.MethodGet, "/", SinglePageApplication(cfg.UI))
	app.Raw(http.MethodGet, "/endpoints/:key", SinglePageApplication(cfg.UI))
	app.Raw(http.MethodGet, "/suites/:key", SinglePageApplication(cfg.UI))
	// The brand's marks, and the manifest that names them. Registered before the
	// static middleware so they answer the bare root paths a browser guesses at.
	RegisterBrandIcons(app, cfg.UI)
	app.Raw(http.MethodGet, "/manifest.json", Manifest(cfg.UI))
	// Health endpoint
	healthHandler := health.Handler().WithJSON(true)
	app.Raw(http.MethodGet, "/health", func(c *zip.Ctx) error {
		statusCode, body := healthHandler.GetResponseStatusCodeAndBody()
		return c.Bytes(statusCode, body)
	})
	// Custom CSS
	app.Raw(http.MethodGet, "/css/custom.css", CustomCSSHandler{customCSS: cfg.UI.CustomCSS}.GetCustomCSS)
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
		if err := cfg.Security.RegisterHandlers(app.Group("").Group("")); err != nil {
			panic(err)
		}
		if err := cfg.Security.ApplySecurityMiddleware(protectedAPIRouter); err != nil {
			panic(err)
		}
	}
	protectedAPIRouter.Raw(http.MethodGet, "/endpoints/statuses", EndpointStatuses(cfg))
	protectedAPIRouter.Raw(http.MethodGet, "/endpoints/:key/statuses", EndpointStatus(cfg))
	protectedAPIRouter.Raw(http.MethodGet, "/suites/statuses", SuiteStatuses(cfg))
	protectedAPIRouter.Raw(http.MethodGet, "/suites/:key/statuses", SuiteStatus(cfg))
	return app
}

// scrape answers a Prometheus scrape from the default gatherer, and counts the
// scrapes it serves on reg under the names promhttp publishes. metric renders
// the exposition as a value — status, headers, body — and this writes those
// three fields the way zip writes a response. The deadline the scraper asked
// for comes off the request's own headers.
func scrape(reg metric.Registerer) zip.Handler {
	served := reg.NewCounter("promhttp_metric_handler_requests_total", "Total number of scrapes served.")
	inFlight := reg.NewGauge("promhttp_metric_handler_requests_in_flight", "Current number of scrapes being served.")
	return func(c *zip.Ctx) error {
		inFlight.Inc()
		defer inFlight.Dec()
		served.Inc()
		e := metric.Scrape(c.Context(), metric.DefaultGatherer, metric.HandlerOpts{DisableCompression: true}, metric.ScrapeTimeout(c.Header))
		for name, value := range e.Header {
			c.SetHeader(name, value)
		}
		return c.Bytes(e.Status, e.Body)
	}
}
