package api

import (
	_ "embed"
	"html/template"

	"github.com/TwiN/logr"
	"github.com/hanzoai/status/config/ui"
	static "github.com/hanzoai/status/web"
	"github.com/zap-proto/zip"
)

func SinglePageApplication(uiConfig *ui.Config) zip.Handler {
	return func(c *zip.Ctx) error {
		vd := ui.ViewData{UI: uiConfig}
		{
			themeFromCookie := c.Fiber().Cookies("theme")
			if len(themeFromCookie) > 0 {
				if themeFromCookie == "dark" {
					vd.Theme = "dark"
				}
			} else if uiConfig.IsDarkMode() { // Since there's no theme cookie, we'll rely on ui.DarkMode
				vd.Theme = "dark"
			}
		}
		t, err := template.ParseFS(static.FileSystem, static.IndexPath)
		if err != nil {
			// This should never happen, because ui.ValidateAndSetDefaults validates that the template works.
			logr.Errorf("[api.SinglePageApplication] Failed to parse template. This should never happen, because the template is validated on start. Error: %s", err.Error())
			return c.String(500, "Failed to parse template. This should never happen, because the template is validated on start.")
		}
		c.SetHeader("Content-Type", "text/html")
		err = t.Execute(c.Fiber(), vd)
		if err != nil {
			// This should never happen, because ui.ValidateAndSetDefaults validates that the template works.
			logr.Errorf("[api.SinglePageApplication] Failed to execute template. This should never happen, because the template is validated on start. Error: %s", err.Error())
			return c.String(500, "Failed to parse template. This should never happen, because the template is validated on start.")
		}
		return c.Fiber().SendStatus(200)
	}
}
