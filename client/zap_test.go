package client

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zap-proto/zip"
)

// TestPerformZAPCheck drives a REAL zip peer over a real unix socket. The two
// cases pin the probe in both directions, so neither a stuck `true` nor a
// blanket `false` on error survives: a peer that is there answers even though
// it serves no probeOp, and an address with nothing behind it does not.
func TestPerformZAPCheck(t *testing.T) {
	dir, err := os.MkdirTemp("", "zp")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "peer.sock")

	// A peer serving ONE op, deliberately not probeOp — the shape every real
	// peer has, since nothing in the fleet implements status.probe.
	app := zip.New(zip.Config{})
	zip.Post(app, "/v1/noop", func(_ context.Context, _ *struct{}) (*struct{}, error) {
		return nil, nil
	}, zip.WithOperationID("noop"))
	go func() { _ = app.Listen(sock) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	for i := 0; ; i++ {
		if c, derr := net.Dial("unix", sock); derr == nil {
			_ = c.Close()
			break
		}
		if i == 100 {
			t.Fatalf("%s never began listening", sock)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// A REFUSAL IS AN ANSWER: this peer has no status.probe, so it refuses —
	// and the refusal completed a round trip, which is what reachable means.
	connected, status, err, d := PerformZAPCheck("zap://"+sock, GetDefaultConfig())
	if !connected {
		t.Fatalf("a live peer refusing an unknown op must read reachable; got err=%v", err)
	}
	if status != "SERVING" || err != nil {
		t.Fatalf("status=%q err=%v", status, err)
	}
	if d <= 0 {
		t.Fatal("a completed round trip must report a duration")
	}

	// Nothing behind the address: no round trip, so no evidence, so not connected.
	connected, _, err, _ = PerformZAPCheck("zap://"+filepath.Join(dir, "absent.sock"), GetDefaultConfig())
	if connected {
		t.Fatal("an address with no peer must not read reachable")
	}
	if err == nil {
		t.Fatal("an unreachable peer must report why")
	}
}
