package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hanzoai/status/config"
	"github.com/hanzoai/status/config/ui"
	static "github.com/hanzoai/status/web"
)

// TestRootIconsServeTheNamedBrand walks the paths a browser reaches for without
// being told and compares the bytes it gets back to the brand's own file. Status
// codes would have passed before this change too: the root paths answered 200 the
// whole time, with the wrong company's mark behind them.
func TestRootIconsServeTheNamedBrand(t *testing.T) {
	api := newTestAPI()
	for rootPath, brandPath := range rootIcons(newTestUIConfig().Icons) {
		t.Run(rootPath, func(t *testing.T) {
			served := get(t, api, rootPath)
			want, err := static.FileSystem.ReadFile(static.RootPath + brandPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(served, want) {
				t.Errorf("%s served %d bytes; %s is %d — the root path is not the brand's file", rootPath, len(served), brandPath, len(want))
			}
		})
	}
}

// TestManifestFollowsTheConfig pins what an installed status page calls itself.
// The static file this replaced said "Status" on a page titled "Hanzo Status",
// described itself in upstream Gatus's words, and splashed #f7f9fb light grey in
// front of a near-black UI — three facts the page already knew, kept in a second
// place and wrong in all three.
func TestManifestFollowsTheConfig(t *testing.T) {
	uiConfig := &ui.Config{
		Brand:       "lux",
		Title:       "Lux Status",
		Description: "Real-time health of the Lux network.",
	}
	if err := uiConfig.ValidateAndSetDefaults(); err != nil {
		t.Fatal(err)
	}
	api := New(&config.Config{UI: uiConfig})
	var got manifest
	if err := json.Unmarshal(get(t, api, "/manifest.json"), &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != uiConfig.Title || got.ShortName != uiConfig.Title {
		t.Errorf("manifest is named %q/%q, want the page's own title %q", got.Name, got.ShortName, uiConfig.Title)
	}
	if got.Description != uiConfig.Description {
		t.Errorf("manifest describes itself as %q, want %q", got.Description, uiConfig.Description)
	}
	if got.BackgroundColor != uiConfig.BackgroundColor() || got.ThemeColor != uiConfig.BackgroundColor() {
		t.Errorf("manifest paints %q/%q, want the colour the app paints, %q", got.BackgroundColor, got.ThemeColor, uiConfig.BackgroundColor())
	}
	if len(got.Icons) != 2 {
		t.Fatalf("expected the 192 and 512 install icons, got %d", len(got.Icons))
	}
	for _, icon := range got.Icons {
		if !strings.HasPrefix(icon.Src, "/brands/lux/") {
			t.Errorf("a Lux status page would install %q", icon.Src)
		}
	}
}

func get(t *testing.T, api *API, path string) []byte {
	t.Helper()
	response, err := api.Router().Fiber().Test(httptest.NewRequest(http.MethodGet, path, http.NoBody))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned %d", path, response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
