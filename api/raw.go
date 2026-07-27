package api

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/hanzoai/status/storage/store"
	"github.com/hanzoai/status/storage/store/common"
	"github.com/zap-proto/zip"
)

func UptimeRaw(c *zip.Ctx) error {
	duration := c.Param("duration")
	var from time.Time
	switch duration {
	case "30d":
		from = time.Now().Add(-30 * 24 * time.Hour)
	case "7d":
		from = time.Now().Add(-7 * 24 * time.Hour)
	case "24h":
		from = time.Now().Add(-24 * time.Hour)
	case "1h":
		from = time.Now().Add(-2 * time.Hour) // Because uptime metrics are stored by hour, we have to cheat a little
	default:
		return c.String(400, "Durations supported: 30d, 7d, 24h, 1h")
	}
	key, err := url.QueryUnescape(c.Param("key"))
	if err != nil {
		return c.String(400, "invalid key encoding")
	}
	uptime, err := store.Get().GetUptimeByKey(key, from, time.Now())
	if err != nil {
		if errors.Is(err, common.ErrEndpointNotFound) {
			return c.String(404, err.Error())
		} else if errors.Is(err, common.ErrInvalidTimeRange) {
			return c.String(400, err.Error())
		}
		return c.String(500, err.Error())
	}

	c.SetHeader("Content-Type", "text/plain")
	c.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
	c.SetHeader("Expires", "0")
	return c.Bytes(200, []byte(fmt.Sprintf("%f", uptime)))
}

func ResponseTimeRaw(c *zip.Ctx) error {
	duration := c.Param("duration")
	var from time.Time
	switch duration {
	case "30d":
		from = time.Now().Add(-30 * 24 * time.Hour)
	case "7d":
		from = time.Now().Add(-7 * 24 * time.Hour)
	case "24h":
		from = time.Now().Add(-24 * time.Hour)
	case "1h":
		from = time.Now().Add(-2 * time.Hour) // Because uptime metrics are stored by hour, we have to cheat a little
	default:
		return c.String(400, "Durations supported: 30d, 7d, 24h, 1h")
	}
	key, err := url.QueryUnescape(c.Param("key"))
	if err != nil {
		return c.String(400, "invalid key encoding")
	}
	responseTime, err := store.Get().GetAverageResponseTimeByKey(key, from, time.Now())
	if err != nil {
		if errors.Is(err, common.ErrEndpointNotFound) {
			return c.String(404, err.Error())
		} else if errors.Is(err, common.ErrInvalidTimeRange) {
			return c.String(400, err.Error())
		}
		return c.String(500, err.Error())
	}

	c.SetHeader("Content-Type", "text/plain")
	c.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
	c.SetHeader("Expires", "0")
	return c.Bytes(200, []byte(fmt.Sprintf("%d", responseTime)))
}
