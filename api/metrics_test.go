package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"hanzo.ai/status/config"
)

// scrapeCounter is the series the metrics route keeps about itself: one count
// per scrape served, whatever the answer was.
const scrapeCounter = "promhttp_metric_handler_requests_total"

// ask sends a scrape to the metrics route and returns what it answered.
func ask(t *testing.T, api *API, header map[string]string) (int, http.Header, string) {
	t.Helper()
	request := httptest.NewRequest("GET", "/metrics", http.NoBody)
	for name, value := range header {
		request.Header.Set(name, value)
	}
	response, err := api.Router().Fiber().Test(request)
	if err != nil {
		t.Fatalf("GET /metrics: %s", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("reading the exposition: %s", err)
	}
	return response.StatusCode, response.Header, string(body)
}

// scrapeCount reads the scrape count out of an exposition.
func scrapeCount(t *testing.T, body string) float64 {
	t.Helper()
	for _, line := range strings.Split(body, "\n") {
		name, value, found := strings.Cut(line, " ")
		if !found || name != scrapeCounter {
			continue
		}
		count, err := strconv.ParseFloat(value, 64)
		if err != nil {
			t.Fatalf("%s is %q, which is not a number", scrapeCounter, value)
		}
		return count
	}
	t.Fatalf("no %s in the exposition:\n%s", scrapeCounter, body)
	return 0
}

// TestMetricsExposition pins what GET /metrics answers, so the route keeps
// answering it however it is written. The version rides in the media type and a
// scraper reads it to choose a parser, so the type is pinned whole rather than
// by prefix. The deadline a scraper asks for is part of the answer: it arrives
// under either spelling, a value that is not a number means no deadline, and a
// deadline already spent is refused rather than answered with a partial body.
func TestMetricsExposition(t *testing.T) {
	api := New(&config.Config{Metrics: true, UI: newTestUIConfig()})

	served := []struct {
		name   string
		header map[string]string
	}{
		{"no deadline asked for", nil},
		{"a deadline", map[string]string{"X-Scrape-Timeout-Seconds": "5"}},
		{"a deadline, spelled the older way", map[string]string{"X-Prometheus-Scrape-Timeout-Seconds": "5"}},
		{"a deadline that is not a number", map[string]string{"X-Scrape-Timeout-Seconds": "soon"}},
	}
	for _, s := range served {
		t.Run(s.name, func(t *testing.T) {
			code, header, body := ask(t, api, s.header)

			if code != http.StatusOK {
				t.Errorf("answered %d, want %d", code, http.StatusOK)
			}
			if got, want := header.Get("Content-Type"), "text/plain; version=0.0.4; charset=utf-8"; got != want {
				t.Errorf("Content-Type is %q, want %q", got, want)
			}
			for _, line := range []string{
				"# HELP " + scrapeCounter + " Total number of scrapes served.",
				"# TYPE " + scrapeCounter + " counter",
				"# TYPE promhttp_metric_handler_requests_in_flight gauge",
			} {
				if !strings.Contains(body, line) {
					t.Errorf("no %q in the exposition:\n%s", line, body)
				}
			}
		})
	}

	t.Run("a deadline already spent", func(t *testing.T) {
		code, header, body := ask(t, api, map[string]string{"X-Scrape-Timeout-Seconds": "1e-9"})

		if code != http.StatusInternalServerError {
			t.Errorf("answered %d, want %d", code, http.StatusInternalServerError)
		}
		if got, want := header.Get("Content-Type"), "text/plain; charset=utf-8"; got != want {
			t.Errorf("Content-Type is %q, want %q", got, want)
		}
		if got, want := header.Get("X-Content-Type-Options"), "nosniff"; got != want {
			t.Errorf("X-Content-Type-Options is %q, want %q", got, want)
		}
		if got, want := body, "metrics gather error\n"; got != want {
			t.Errorf("body is %q, want %q", got, want)
		}
	})

	t.Run("every scrape is counted", func(t *testing.T) {
		_, _, first := ask(t, api, nil)
		_, _, second := ask(t, api, nil)

		if before, after := scrapeCount(t, first), scrapeCount(t, second); after != before+1 {
			t.Errorf("%s went %v -> %v, want one more", scrapeCounter, before, after)
		}
	})
}
