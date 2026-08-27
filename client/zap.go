package client

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/zap-proto/zip"
)

// probeOp is the op a reachability check asks for. No peer is expected to serve
// it — the question is whether one ANSWERS, and a name nobody implements is the
// only one that cannot have a side effect on the peer being probed.
const probeOp = "status.probe"

// PerformZAPCheck dials a ZAP peer and reports whether it answers.
//
// THERE IS NO HEALTH OP TO CALL, and that is a decision zip states rather than
// an omission. Its health surface is /healthz on a second listener and it is
// HTTP on purpose: "a probe against a ZAP socket is read as a frame — 'GET '
// arrives as frame size 1195725856" (zip ops.go). So an endpoint that wants
// /healthz is an HTTP endpoint and already has a type here. What only ZAP can
// answer is whether the PEER IS THERE, which is the half a gRPC probe's
// connection state used to carry and the half no HTTP GET reaches on a socket
// that speaks frames.
//
// A REFUSAL IS AN ANSWER. zip reports a transport failure as its own 502
// (Errorf(502, "zip: call %s at %s")) and rebuilds a callee's refusal with the
// status the callee chose (remoteError), so an *HTTPError carrying any other
// status came back over a COMPLETED round trip — the dial, the handshake and
// the framing all worked, which is exactly what a reachability check asks. A
// 502 is the one status both cases can produce, so it is never read as
// evidence: with nothing positive to go on a monitor reports not connected
// rather than guessing, and an unreachable peer is never reported up.
func PerformZAPCheck(address string, cfg *Config) (bool, string, error, time.Duration) {
	if cfg == nil {
		cfg = GetDefaultConfig()
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	start := time.Now()
	conn, err := zip.Dial(strings.TrimPrefix(address, "zap://"))
	if err != nil {
		return false, "", err, time.Since(start)
	}
	defer conn.Close()

	_, err = zip.Call[struct{}, struct{}](ctx, conn, probeOp, nil)
	duration := time.Since(start)
	if err == nil {
		return true, "SERVING", nil, duration
	}
	var he *zip.HTTPError
	if errors.As(err, &he) && he.Status != 502 {
		// The peer chose this status, so the peer is there.
		return true, "SERVING", nil, duration
	}
	return false, "", err, duration
}
