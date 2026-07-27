// Package zipx is the single bridge between fiber-native middleware and zip's
// handler signature.
//
// zip owns the handler type (func(*zip.Ctx) error) and ships adapters for
// net/http (zip.AdaptNetHTTP*), but the fiber middleware catalogue it is built
// on speaks fiber.Handler. Wrap is the one place that crossing happens, so no
// call site has to reach for c.Fiber() itself.
package zipx

import (
	fiber "github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/zip"
)

// Wrap adapts a fiber.Handler — any fiber v3 middleware — to a zip.Handler so
// it registers with app.Use / router.Use like native zip middleware.
func Wrap(h fiber.Handler) zip.Handler {
	return func(c *zip.Ctx) error { return h(c.Fiber()) }
}
