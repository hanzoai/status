package endpoint

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// api.hanzo.ai serves the embedded console SPA on any path the binary does not
// route, so an unrouted path answers 200 and a status-code-only condition
// cannot distinguish a served route from the SPA catch-all. Every row therefore
// asserts on the BODY.
//
// This test evaluates the board's REAL conditions against the LIVE responses
// using the same evaluator the status binary runs, so "the conditions are
// correct" is measured rather than argued — a typo'd JSONPath or an unsupported
// function turns the whole board red in production, and that is not something
// to discover from the board.
//
//	HANZO_BOARD=/path/to/configmap-hanzo.yaml go test ./config/endpoint -run TestHanzoBoard -v
//
// Skipped unless HANZO_BOARD is set: it needs the network, so it must never
// gate an offline build.
func TestHanzoBoardConditionsHoldLive(t *testing.T) {
	path := os.Getenv("HANZO_BOARD")
	if path == "" {
		t.Skip("set HANZO_BOARD to the status ConfigMap to run the live board check")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read board: %v", err)
	}
	var cm struct {
		Data struct {
			Config string `yaml:"config.yaml"`
		} `yaml:"data"`
	}
	if err := yaml.Unmarshal(raw, &cm); err != nil {
		t.Fatalf("parse ConfigMap: %v", err)
	}
	var board struct {
		Endpoints []struct {
			Name       string      `yaml:"name"`
			Group      string      `yaml:"group"`
			URL        string      `yaml:"url"`
			Conditions []Condition `yaml:"conditions"`
		} `yaml:"endpoints"`
	}
	if err := yaml.Unmarshal([]byte(cm.Data.Config), &board); err != nil {
		t.Fatalf("parse board config: %v", err)
	}
	if len(board.Endpoints) == 0 {
		t.Fatal("board declares no endpoints")
	}

	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}},
	}

	var checked, skipped, withBody int
	for _, ep := range board.Endpoints {
		t.Run(ep.Group+"/"+ep.Name, func(t *testing.T) {
			// Every condition must PARSE, whether or not the URL is reachable
			// from here. This is the half that catches an unsupported function.
			for _, c := range ep.Conditions {
				if err := c.Validate(); err != nil {
					t.Errorf("condition %q does not validate: %v", c, err)
				}
				if strings.Contains(string(c), "[BODY]") {
					withBody++
				}
			}
			// An in-cluster Service is not resolvable from a dev host; parse-only.
			if strings.Contains(ep.URL, ".svc:") || strings.Contains(ep.URL, ".svc.") {
				skipped++
				t.Logf("in-cluster URL, conditions parsed but not evaluated: %s", ep.URL)
				return
			}

			start := time.Now()
			resp, err := client.Get(ep.URL)
			if err != nil {
				t.Fatalf("GET %s: %v", ep.URL, err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			result := &Result{
				HTTPStatus: resp.StatusCode,
				Body:       body,
				Duration:   time.Since(start),
			}
			checked++
			for _, c := range ep.Conditions {
				// Response time is the machine's, not the cluster's — a dev host
				// over the public internet is not the probe's vantage point, so
				// asserting it here would measure the wrong thing.
				if strings.Contains(string(c), "[RESPONSE_TIME]") {
					continue
				}
				if !c.evaluate(result, false, false, nil) {
					preview := string(body)
					if len(preview) > 160 {
						preview = preview[:160]
					}
					t.Errorf("condition FAILED: %s\n  url:    %s\n  status: %d\n  body:   %s",
						c, ep.URL, resp.StatusCode, preview)
				}
			}
		})
	}
	t.Logf("evaluated %d endpoints live (%d in-cluster parse-only), %d body conditions",
		checked, skipped, withBody)
}
