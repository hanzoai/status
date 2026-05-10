//go:build !grpc

package client

import (
	"errors"
	"time"
)

var errGRPCDisabled = errors.New("status: gRPC health check disabled — rebuild with -tags grpc to enable, or use the ZAP-native endpoint type")

func PerformGRPCHealthCheck(address string, useTLS bool, cfg *Config) (bool, string, error, time.Duration) {
	return false, "", errGRPCDisabled, 0
}
