package api

import (
	"encoding/json"
	"path"

	"github.com/TwiN/logr"
	"github.com/hanzoai/status/config/ui"
	static "github.com/hanzoai/status/web"
	"github.com/zap-proto/zip"
)

// rootIcons are the paths a browser reaches for on its own, without being told:
// /favicon.ico when no <link rel=icon> matched, /apple-touch-icon.png when iOS
// saves a home screen shortcut, the manifest icons when a PWA is installed.
//
// They used to be real files sitting at the web root. One binary serves every
// brand we run, so whichever brand's bytes got copied to the root were handed to
// all the others: status.lux.network's home screen icon was the Hanzo mark, byte
// for byte identical. Artwork has one home now — /brands/<brand>/ — and these are
// views onto it rather than a second copy that drifts.
func rootIcons(icons ui.Icons) map[string]string {
	return map[string]string{
		"/favicon.ico":          icons.ICO,
		"/favicon.svg":          icons.SVG,
		"/favicon-16x16.png":    icons.PNG16,
		"/favicon-32x32.png":    icons.PNG32,
		"/apple-touch-icon.png": icons.Touch180,
		"/logo-192x192.png":     icons.PNG192,
		"/logo-512x512.png":     icons.PNG512,
	}
}

// RegisterBrandIcons serves each bare root path from the configured brand's
// directory. The bytes are read once at startup: ui.Config.ValidateAndSetDefaults
// has already proven every one of them is in the binary, so a brand that cannot
// be served is a boot failure rather than a page wearing nobody's mark.
func RegisterBrandIcons(app *zip.App, uiConfig *ui.Config) {
	for rootPath, brandPath := range rootIcons(uiConfig.Icons) {
		if len(brandPath) == 0 { // no brand named, so there is no mark to hand out
			continue
		}
		body, err := static.FileSystem.ReadFile(static.RootPath + brandPath)
		if err != nil {
			logr.Errorf("[api.RegisterBrandIcons] %s is missing from the binary: %s", brandPath, err.Error())
			continue
		}
		contentType := iconContentType(brandPath)
		app.Get(rootPath, func(c *zip.Ctx) error {
			c.SetHeader("Content-Type", contentType)
			c.SetHeader("Cache-Control", "public, max-age=3600")
			return c.Bytes(200, body)
		})
	}
}

func iconContentType(p string) string {
	switch path.Ext(p) {
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "image/png"
	}
}

// manifest is the PWA manifest: what an installed status page calls itself, and
// what it paints before it has painted anything.
type manifest struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	ShortName       string         `json:"short_name"`
	Description     string         `json:"description"`
	Lang            string         `json:"lang"`
	Scope           string         `json:"scope"`
	StartURL        string         `json:"start_url"`
	ThemeColor      string         `json:"theme_color"`
	BackgroundColor string         `json:"background_color"`
	Display         string         `json:"display"`
	Icons           []manifestIcon `json:"icons"`
}

type manifestIcon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

// Manifest renders the manifest from the same ui.Config that renders the page.
// It was a checked-in static file, which is how an installed "Hanzo Status" came
// to call itself "Status", describe itself in upstream Gatus's words, and splash
// #f7f9fb light grey in front of a page that paints near-black. Three facts that
// the page already knew, kept in a second place, all three wrong.
func Manifest(uiConfig *ui.Config) zip.Handler {
	background := uiConfig.BackgroundColor()
	m := manifest{
		ID:              "/",
		Name:            uiConfig.Title,
		ShortName:       uiConfig.Title,
		Description:     uiConfig.Description,
		Lang:            "en",
		Scope:           "/",
		StartURL:        "/",
		ThemeColor:      background,
		BackgroundColor: background,
		Display:         "standalone",
		Icons:           []manifestIcon{},
	}
	if len(uiConfig.Icons.PNG192) > 0 {
		m.Icons = append(m.Icons,
			manifestIcon{Src: uiConfig.Icons.PNG192, Sizes: "192x192", Type: "image/png"},
			manifestIcon{Src: uiConfig.Icons.PNG512, Sizes: "512x512", Type: "image/png"},
		)
	}
	body, err := json.Marshal(m)
	if err != nil {
		// Unreachable: every field is a string or a slice of strings.
		logr.Errorf("[api.Manifest] Failed to marshal manifest: %s", err.Error())
	}
	return func(c *zip.Ctx) error {
		c.SetHeader("Content-Type", "application/manifest+json")
		return c.Bytes(200, body)
	}
}
