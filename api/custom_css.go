package api

import (
	"github.com/zap-proto/zip"
)

type CustomCSSHandler struct {
	customCSS string
}

func (handler CustomCSSHandler) GetCustomCSS(c *zip.Ctx) error {
	c.SetHeader("Content-Type", "text/css")
	return c.String(200, handler.customCSS)
}
