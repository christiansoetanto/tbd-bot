package util_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/christiansoetanto/tbd-bot/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// recorder captures what the heartbeat actually sent, since the whole point of
// the switch is which URL gets hit rather than what the bot believes it did.
type recorder struct {
	mu     sync.Mutex
	paths  []string
	bodies []string
	status int
}

func (r *recorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	r.mu.Lock()
	r.paths = append(r.paths, req.URL.Path)
	r.bodies = append(r.bodies, string(body))
	status := r.status
	r.mu.Unlock()
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
}

func (r *recorder) calls() ([]string, []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.paths...), append([]string(nil), r.bodies...)
}

func gaugeValue(t *testing.T, name string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() != name {
			continue
		}
		metrics := mf.GetMetric()
		if len(metrics) == 0 {
			t.Fatalf("metric %s has no samples", name)
		}
		return metrics[0].GetGauge().GetValue()
	}
	t.Fatalf("metric %s not found in default gatherer", name)
	return 0
}

// A deploy takes about 26s before the gateway is up, so a tick landing in that
// window would ping /fail and alert on every single deploy. Alerts that cry
// wolf on every deploy are alerts nobody reads.
func TestHeartbeatStaysSilentUntilGatewayEverConnects(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	rec := &recorder{}
	server := httptest.NewServer(rec)
	t.Cleanup(server.Close)

	util.SetGatewayStater(fakeStater{}) // registered, never acked
	util.NewHeartbeat(server.URL).Ping(context.Background())

	if paths, _ := rec.calls(); len(paths) != 0 {
		t.Errorf("expected no ping before the gateway has ever connected, got %v", paths)
	}
}

func TestHeartbeatPingsSuccessURLWhenHealthy(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	rec := &recorder{}
	server := httptest.NewServer(rec)
	t.Cleanup(server.Close)

	util.SetGatewayStater(fakeStater{ack: time.Now()})
	before := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("success"))
	util.NewHeartbeat(server.URL).Ping(context.Background())

	paths, _ := rec.calls()
	if len(paths) != 1 {
		t.Fatalf("expected exactly one ping, got %v", paths)
	}
	if strings.HasSuffix(paths[0], "/fail") {
		t.Errorf("expected the success URL while healthy, got %q", paths[0])
	}
	if after := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("success")); after-before != 1 {
		t.Errorf("expected result=success to increment by 1, before=%f after=%f", before, after)
	}
}

// The reason travels in the body so the Discord message says what is wrong
// rather than just that a check failed.
func TestHeartbeatPingsFailURLWithReasonWhenStale(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	rec := &recorder{}
	server := httptest.NewServer(rec)
	t.Cleanup(server.Close)

	util.SetGatewayStater(fakeStater{ack: time.Now().Add(-util.GatewayStaleAfter - time.Minute)})
	before := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("fail"))
	util.NewHeartbeat(server.URL).Ping(context.Background())

	paths, bodies := rec.calls()
	if len(paths) != 1 {
		t.Fatalf("expected exactly one ping, got %v", paths)
	}
	if !strings.HasSuffix(paths[0], "/fail") {
		t.Errorf("expected the /fail URL while stale, got %q", paths[0])
	}
	if !strings.Contains(bodies[0], "heartbeat") {
		t.Errorf("expected the unhealthy reason in the body, got %q", bodies[0])
	}
	if after := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("fail")); after-before != 1 {
		t.Errorf("expected result=fail to increment by 1, before=%f after=%f", before, after)
	}
}

// A trailing slash in the pasted URL must not produce //fail.
func TestHeartbeatFailURLHandlesTrailingSlash(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	rec := &recorder{}
	server := httptest.NewServer(rec)
	t.Cleanup(server.Close)

	util.SetGatewayStater(fakeStater{ack: time.Now().Add(-util.GatewayStaleAfter - time.Minute)})
	util.NewHeartbeat(server.URL + "/").Ping(context.Background())

	paths, _ := rec.calls()
	if len(paths) != 1 {
		t.Fatalf("expected exactly one ping, got %v", paths)
	}
	if strings.Contains(paths[0], "//") {
		t.Errorf("expected a single slash before fail, got %q", paths[0])
	}
}

// This is the GRAFANA_DISCORD_WEBHOOK_URL trap: an unset URL provisions
// cleanly, reads green everywhere, and delivers nothing. The gauge is what
// lets Grafana alert on it.
func TestHeartbeatUnsetURLIsDisabledAndVisible(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	util.SetGatewayStater(fakeStater{ack: time.Now()})
	hb := util.NewHeartbeat("")
	if hb.Enabled() {
		t.Error("expected the heartbeat to be disabled when the URL is unset")
	}
	hb.Ping(context.Background())

	if got := gaugeValue(t, "tbd_bot_external_heartbeat_enabled"); got != 0 {
		t.Errorf("expected tbd_bot_external_heartbeat_enabled to be 0 when unset, got %f", got)
	}
}

func TestHeartbeatEnabledGaugeIsOneWhenConfigured(t *testing.T) {
	util.NewHeartbeat("https://hc-ping.example/abc")
	if got := gaugeValue(t, "tbd_bot_external_heartbeat_enabled"); got != 1 {
		t.Errorf("expected tbd_bot_external_heartbeat_enabled to be 1 when configured, got %f", got)
	}
	util.NewHeartbeat("")
}

// A ping that never lands must not look like a delivered one, or the metric
// that proves the pinger works would itself lie.
func TestHeartbeatCountsTransportFailureAsError(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	server := httptest.NewServer(&recorder{})
	url := server.URL
	server.Close() // nothing is listening now

	util.SetGatewayStater(fakeStater{ack: time.Now()})
	before := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("error"))
	util.NewHeartbeat(url).Ping(context.Background())

	if after := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("error")); after-before != 1 {
		t.Errorf("expected result=error to increment by 1, before=%f after=%f", before, after)
	}
}

// A wrong UUID answers 404. That is a broken switch, not a delivered ping.
func TestHeartbeatCountsNon2xxAsError(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	rec := &recorder{status: http.StatusNotFound}
	server := httptest.NewServer(rec)
	t.Cleanup(server.Close)

	util.SetGatewayStater(fakeStater{ack: time.Now()})
	beforeError := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("error"))
	beforeSuccess := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("success"))
	util.NewHeartbeat(server.URL).Ping(context.Background())

	if after := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("error")); after-beforeError != 1 {
		t.Errorf("expected a 404 to count as error, before=%f after=%f", beforeError, after)
	}
	if after := testutil.ToFloat64(util.ExternalHeartbeatTotal.WithLabelValues("success")); after != beforeSuccess {
		t.Errorf("expected a 404 not to count as success, before=%f after=%f", beforeSuccess, after)
	}
}

// The gauge has to track any delivered ping, not only healthy ones. Tracking
// success alone would make a gateway outage also read as a broken pinger, and
// group_by alertname would then send that outage to Discord twice.
func TestHeartbeatLastPingGaugeTracksDeliveredFailPings(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	rec := &recorder{}
	server := httptest.NewServer(rec)
	t.Cleanup(server.Close)

	util.SetGatewayStater(fakeStater{ack: time.Now().Add(-util.GatewayStaleAfter - time.Minute)})
	util.NewHeartbeat(server.URL).Ping(context.Background())

	got := gaugeValue(t, "tbd_bot_external_heartbeat_last_ping_timestamp_seconds")
	if age := time.Since(time.Unix(int64(got), 0)); age > time.Minute {
		t.Errorf("expected the last ping timestamp to be current after a delivered /fail, got %s ago", age)
	}
}

func TestGatewayEverConnected(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	util.SetGatewayStater(nil)
	if util.GatewayEverConnected() {
		t.Error("expected false with no stater registered")
	}
	util.SetGatewayStater(fakeStater{})
	if util.GatewayEverConnected() {
		t.Error("expected false before the first ack")
	}
	util.SetGatewayStater(fakeStater{ack: time.Now().Add(-time.Hour)})
	if !util.GatewayEverConnected() {
		t.Error("expected true once an ack has been seen, even a stale one")
	}
}

// The ping inherits whatever context Init was handed. That is context.Background()
// today, but a startup deadline added there later would expire and every
// subsequent ping would fail with "context deadline exceeded" — the switch
// would go silent and alert about a bot that is fine. The heartbeat owns its
// own deadline instead.
func TestHeartbeatIgnoresParentContextLifetime(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	rec := &recorder{}
	server := httptest.NewServer(rec)
	t.Cleanup(server.Close)

	expired, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer cancel()

	util.SetGatewayStater(fakeStater{ack: time.Now()})
	util.NewHeartbeat(server.URL).Ping(expired)

	if paths, _ := rec.calls(); len(paths) != 1 {
		t.Fatalf("expected the ping to land despite an expired parent context, got %v", paths)
	}
}
