package api

import (
	"errors"
	"strings"
	"time"

	"github.com/TwiN/logr"
	"github.com/hanzoai/status/config"
	"github.com/hanzoai/status/config/endpoint"
	"github.com/hanzoai/status/metrics"
	"github.com/hanzoai/status/storage/store"
	"github.com/hanzoai/status/storage/store/common"
	"github.com/hanzoai/status/watchdog"
	"github.com/zap-proto/zip"
)

func CreateExternalEndpointResult(cfg *config.Config) zip.Handler {
	extraLabels := cfg.GetUniqueExtraMetricLabels()
	return func(c *zip.Ctx) error {
		// Check if the success query parameter is present
		success, exists := c.Fiber().Queries()["success"]
		if !exists || (success != "true" && success != "false") {
			return c.String(400, "missing or invalid success query parameter")
		}
		// Check if the authorization bearer token header is correct
		authorizationHeader := c.Header("Authorization")
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			return c.String(401, "invalid Authorization header")
		}
		token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
		if len(token) == 0 {
			return c.String(401, "bearer token must not be empty")
		}
		key := c.Param("key")
		externalEndpoint := cfg.GetExternalEndpointByKey(key)
		if externalEndpoint == nil {
			logr.Errorf("[api.CreateExternalEndpointResult] External endpoint with key=%s not found", key)
			return c.String(404, "not found")
		}
		if externalEndpoint.Token != token {
			logr.Errorf("[api.CreateExternalEndpointResult] Invalid token for external endpoint with key=%s", key)
			return c.String(401, "invalid token")
		}
		// Persist the result in the storage
		result := &endpoint.Result{
			Timestamp: time.Now(),
			Success:   success == "true",
			Errors:    []string{},
		}
		if len(c.Query("duration")) > 0 {
			parsedDuration, err := time.ParseDuration(c.Query("duration"))
			if err != nil {
				logr.Errorf("[api.CreateExternalEndpointResult] Invalid duration from string=%s with error: %s", c.Query("duration"), err.Error())
				return c.String(400, "invalid duration: "+err.Error())
			}
			result.Duration = parsedDuration
		}
		if errorFromQuery := c.Query("error"); !result.Success && len(errorFromQuery) > 0 {
			result.AddError(strings.Clone(errorFromQuery))
		}
		convertedEndpoint := externalEndpoint.ToEndpoint()
		if err := store.Get().InsertEndpointResult(convertedEndpoint, result); err != nil {
			if errors.Is(err, common.ErrEndpointNotFound) {
				return c.String(404, err.Error())
			}
			logr.Errorf("[api.CreateExternalEndpointResult] Failed to insert result in storage: %s", err.Error())
			return c.String(500, err.Error())
		}
		logr.Infof("[api.CreateExternalEndpointResult] Successfully inserted result for external endpoint with key=%s and success=%s", c.Param("key"), success)
		inEndpointMaintenanceWindow := false
		for _, maintenanceWindow := range externalEndpoint.MaintenanceWindows {
			if maintenanceWindow.IsUnderMaintenance() {
				logr.Debug("[api.CreateExternalEndpointResult] Under endpoint maintenance window")
				inEndpointMaintenanceWindow = true
			}
		}
		// Check if an alert should be triggered or resolved
		if !cfg.Maintenance.IsUnderMaintenance() && !inEndpointMaintenanceWindow {
			watchdog.HandleAlerting(convertedEndpoint, result, cfg.Alerting)
			externalEndpoint.NumberOfSuccessesInARow = convertedEndpoint.NumberOfSuccessesInARow
			externalEndpoint.NumberOfFailuresInARow = convertedEndpoint.NumberOfFailuresInARow
		} else {
			logr.Debug("[api.CreateExternalEndpointResult] Not handling alerting because currently in the maintenance window")
		}
		if cfg.Metrics {
			metrics.PublishMetricsForEndpoint(convertedEndpoint, result, extraLabels)
		}
		// Return the result
		return c.String(200, "")
	}
}
