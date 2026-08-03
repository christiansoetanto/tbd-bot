package util_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// fakeStater stands in for *discordgo.Session so gateway liveness can be
// driven to states a real session only reaches during an outage.
type fakeStater struct {
	ack     time.Time
	latency time.Duration
}

func (f fakeStater) LastHeartbeatAck() time.Time     { return f.ack }
func (f fakeStater) HeartbeatLatency() time.Duration { return f.latency }

// The 08-01 outage left the process alive and /metrics answering while the
// gateway was dead, so liveness has to come from a signal that decays on its
// own rather than from a flag some handler is responsible for clearing.
func TestGatewayHealthy_StaleHeartbeatAckIsUnhealthy(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	util.SetGatewayStater(fakeStater{ack: time.Now(), latency: 40 * time.Millisecond})
	if healthy, reason := util.GatewayHealthy(); !healthy {
		t.Errorf("expected healthy with a fresh heartbeat ack, got unhealthy: %s", reason)
	}

	util.SetGatewayStater(fakeStater{ack: time.Now().Add(-util.GatewayStaleAfter - time.Second)})
	healthy, reason := util.GatewayHealthy()
	if healthy {
		t.Errorf("expected unhealthy once the ack is older than %s", util.GatewayStaleAfter)
	}
	if reason == "" {
		t.Error("expected a reason explaining why the gateway is unhealthy")
	}
}

// A gateway that never finishes identifying never acks, so the zero value has
// to read as unhealthy rather than as "no data yet".
func TestGatewayHealthy_ZeroAckIsUnhealthy(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	util.SetGatewayStater(fakeStater{})
	if healthy, _ := util.GatewayHealthy(); healthy {
		t.Error("expected unhealthy when no heartbeat has ever been acked")
	}
}

// Metrics are scraped before main wires the session in, and /health is public,
// so an unset stater must report unhealthy instead of panicking.
func TestGatewayHealthy_NoStaterIsUnhealthyNotPanic(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	util.SetGatewayStater(nil)
	if healthy, _ := util.GatewayHealthy(); healthy {
		t.Error("expected unhealthy when no stater is configured")
	}
	if _, err := prometheus.DefaultGatherer.Gather(); err != nil {
		t.Fatalf("gathering with no stater configured failed: %v", err)
	}
}

func TestGatewayMetricsExposed(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })

	util.SetGatewayStater(fakeStater{ack: time.Unix(1785595555, 0), latency: 250 * time.Millisecond})
	util.SetDiscordConnected(true)
	util.IncGatewayEvent("resumed")

	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	expected := map[string]bool{
		"tbd_bot_discord_connected":                            false,
		"tbd_bot_discord_last_heartbeat_ack_timestamp_seconds": false,
		"tbd_bot_discord_heartbeat_latency_seconds":            false,
		"tbd_bot_discord_gateway_events_total":                 false,
	}
	for _, mf := range metricFamilies {
		if _, ok := expected[mf.GetName()]; ok {
			expected[mf.GetName()] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("metric %s not found in default gatherer", name)
		}
	}

	if got := testutil.ToFloat64(util.DiscordConnected); got != 1 {
		t.Errorf("expected tbd_bot_discord_connected to be 1, got %f", got)
	}
	util.SetDiscordConnected(false)
	if got := testutil.ToFloat64(util.DiscordConnected); got != 0 {
		t.Errorf("expected tbd_bot_discord_connected to be 0, got %f", got)
	}
}

// Status codes go straight into a metric label, so they have to collapse to a
// fixed set. Using the raw code would let a misbehaving upstream grow the
// series count without bound.
func TestHTTPFailureReasonIsBounded(t *testing.T) {
	cases := map[int]string{
		401: "unauthorized",
		403: "forbidden",
		404: "not_found",
		429: "rate_limited",
		418: "client_error",
		503: "server_error",
		200: "unknown",
	}
	for status, want := range cases {
		if got := util.HTTPFailureReason(status); got != want {
			t.Errorf("HTTPFailureReason(%d) = %q, want %q", status, got, want)
		}
	}
}

// The cron jobs only ever see discordgo's error type, so the status code has
// to be recovered from it rather than from a response the caller never holds.
func TestIncDiscordRESTFailureClassifiesRESTError(t *testing.T) {
	before := testutil.ToFloat64(util.ExternalAPIFailuresTotal.WithLabelValues("discord", "unauthorized"))
	util.IncDiscordRESTFailure(&discordgo.RESTError{
		Response: &http.Response{StatusCode: http.StatusUnauthorized},
	})
	if after := testutil.ToFloat64(util.ExternalAPIFailuresTotal.WithLabelValues("discord", "unauthorized")); after-before != 1 {
		t.Errorf("expected a 401 REST error to count as unauthorized, before=%f after=%f", before, after)
	}

	beforeTransport := testutil.ToFloat64(util.ExternalAPIFailuresTotal.WithLabelValues("discord", "transport_error"))
	util.IncDiscordRESTFailure(errors.New("dial tcp: connection refused"))
	if after := testutil.ToFloat64(util.ExternalAPIFailuresTotal.WithLabelValues("discord", "transport_error")); after-beforeTransport != 1 {
		t.Errorf("expected a non-REST error to count as transport_error, before=%f after=%f", beforeTransport, after)
	}
}

// The GitHub token expired on 07-30 and nothing surfaced it for four days
// because no counter recorded the failures.
func TestExternalAPIFailureCounter(t *testing.T) {
	before := testutil.ToFloat64(util.ExternalAPIFailuresTotal.WithLabelValues("github", "unauthorized"))
	util.IncExternalAPIFailure("github", "unauthorized")
	after := testutil.ToFloat64(util.ExternalAPIFailuresTotal.WithLabelValues("github", "unauthorized"))
	if after-before != 1 {
		t.Fatalf("expected the failure counter to increment by 1, got before=%f after=%f", before, after)
	}

	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	for _, mf := range metricFamilies {
		if mf.GetName() == "tbd_bot_external_api_failures_total" {
			return
		}
	}
	t.Error("metric tbd_bot_external_api_failures_total not found in default gatherer")
}
