package api

import (
	"fmt"

	"github.com/hanzoai/status/config"
	"github.com/hanzoai/status/config/suite"
	"github.com/hanzoai/status/storage/store"
	"github.com/hanzoai/status/storage/store/common/paging"
	"github.com/zap-proto/zip"
)

// SuiteStatuses handles requests to retrieve all suite statuses
func SuiteStatuses(cfg *config.Config) zip.Handler {
	return func(c *zip.Ctx) error {
		page, pageSize := extractPageAndPageSizeFromRequest(c, 100)
		params := paging.NewSuiteStatusParams().WithPagination(page, pageSize)
		suiteStatuses, err := store.Get().GetAllSuiteStatuses(params)
		if err != nil {
			return writeJSON(c, 500, map[string]any{
				"error": fmt.Sprintf("Failed to retrieve suite statuses: %v", err),
			})
		}
		// If no statuses exist yet, create empty ones from config
		if len(suiteStatuses) == 0 {
			for _, s := range cfg.Suites {
				if s.IsEnabled() {
					suiteStatuses = append(suiteStatuses, suite.NewStatus(s))
				}
			}
		}
		return writeJSON(c, 200, suiteStatuses)
	}
}

// SuiteStatus handles requests to retrieve a single suite's status
func SuiteStatus(cfg *config.Config) zip.Handler {
	return func(c *zip.Ctx) error {
		page, pageSize := extractPageAndPageSizeFromRequest(c, 100)
		key := c.Param("key")
		params := paging.NewSuiteStatusParams().WithPagination(page, pageSize)
		status, err := store.Get().GetSuiteStatusByKey(key, params)
		if err != nil || status == nil {
			// Try to find the suite in config
			for _, s := range cfg.Suites {
				if s.Key() == key {
					status = suite.NewStatus(s)
					break
				}
			}
			if status == nil {
				return writeJSON(c, 404, map[string]any{
					"error": fmt.Sprintf("Suite with key '%s' not found", key),
				})
			}
		}
		return writeJSON(c, 200, status)
	}
}
