package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/hanzoai/status/config"
	"github.com/hanzoai/status/config/ui"
)

// declaredRoutes is the published HTTP surface of this service, pinned exactly.
// Adding, removing or renaming a route is an API change and must be a
// deliberate edit here — never a side effect of a refactor or a framework move.
var declaredRoutes = []string{
	"GET     /",
	"GET     /css/custom.css",
	"GET     /endpoints/:key",
	"GET     /health",
	"GET     /metrics",
	"GET     /suites/:key",
	"GET     /v1/status/config",
	"GET     /v1/status/endpoints/:key/health/badge.shields",
	"GET     /v1/status/endpoints/:key/health/badge.svg",
	"GET     /v1/status/endpoints/:key/response-times/:duration",
	"GET     /v1/status/endpoints/:key/response-times/:duration/badge.svg",
	"GET     /v1/status/endpoints/:key/response-times/:duration/chart.svg",
	"GET     /v1/status/endpoints/:key/response-times/:duration/history",
	"GET     /v1/status/endpoints/:key/statuses",
	"GET     /v1/status/endpoints/:key/uptimes/:duration",
	"GET     /v1/status/endpoints/:key/uptimes/:duration/badge.svg",
	"GET     /v1/status/endpoints/statuses",
	"GET     /v1/status/suites/:key/statuses",
	"GET     /v1/status/suites/statuses",
	"POST    /v1/status/endpoints/:key/external",
}

func newTestAPI() *API {
	return New(&config.Config{Metrics: true, UI: &ui.Config{}})
}

func TestRouteTable(t *testing.T) {
	routes := newTestAPI().Router().Fiber().GetRoutes(true)
	served := make(map[string]bool, len(routes))
	for _, route := range routes {
		served[route.Method+" "+route.Path] = true
	}
	var actual []string
	for _, route := range routes {
		// The HEAD mirror of a GET is generated, not authored, so it is not part
		// of the declared surface — TestRouteTable_HeadMirrorsGet owns it. zip
		// materialises the mirrors into the route table at build; fiber used to
		// create them lazily, on the first HEAD request, so they were served
		// then too and simply did not show up here. A HEAD with no GET behind it
		// is not a mirror and still has to be declared.
		if route.Method == http.MethodHead && served[http.MethodGet+" "+route.Path] {
			continue
		}
		actual = append(actual, fmt.Sprintf("%-7s %s", route.Method, route.Path))
	}
	sort.Strings(actual)
	if strings.Join(actual, "\n") != strings.Join(declaredRoutes, "\n") {
		t.Errorf("route table changed\nwant:\n%s\n\ngot:\n%s", strings.Join(declaredRoutes, "\n"), strings.Join(actual, "\n"))
	}
}

// TestRouteTable_EveryRouteResolves walks every declared route with a concrete
// value substituted for each parameter and asserts the router dispatched to a
// handler. A handler is free to answer 4xx; what must never happen is the
// router itself missing the path.
func TestRouteTable_EveryRouteResolves(t *testing.T) {
	api := newTestAPI()
	replacer := strings.NewReplacer(":key", "core_frontend", ":duration", "24h")
	for _, declared := range declaredRoutes {
		method, pattern, _ := strings.Cut(declared, " ")
		path := replacer.Replace(strings.TrimSpace(pattern))
		t.Run(method+" "+path, func(t *testing.T) {
			response, err := api.Router().Fiber().Test(httptest.NewRequest(method, path, http.NoBody))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, _ := io.ReadAll(response.Body)
			if strings.HasPrefix(string(body), "Cannot "+method) {
				t.Errorf("%s %s was not routed to a handler: %d %s", method, path, response.StatusCode, body)
			}
		})
	}
}

// TestRouteTable_HeadMirrorsGet pins the auto-generated HEAD mirror of every
// GET route. It is answered either way: zip materialises the mirrors into the
// route table at build, and fiber created them lazily before that.
func TestRouteTable_HeadMirrorsGet(t *testing.T) {
	api := newTestAPI()
	for _, path := range []string{"/", "/health", "/css/custom.css", "/v1/status/config"} {
		response, err := api.Router().Fiber().Test(httptest.NewRequest("HEAD", path, http.NoBody))
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Errorf("HEAD %s should have returned 200, but returned %d instead", path, response.StatusCode)
		}
	}
}

// TestRouteTable_UnknownPathIs404 pins the static-fallback tail of the chain:
// a path that matches no route and no embedded file must 404 rather than be
// swallowed by the catch-all.
func TestRouteTable_UnknownPathIs404(t *testing.T) {
	api := newTestAPI()
	response, err := api.Router().Fiber().Test(httptest.NewRequest("GET", "/v1/status/this-route-does-not-exist", http.NoBody))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("unknown path should have returned 404, but returned %d instead", response.StatusCode)
	}
}
