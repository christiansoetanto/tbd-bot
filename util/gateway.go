package util

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
)

// GatewayStaleAfter is how long the gateway may go without acknowledging a
// heartbeat before it counts as dead. discordgo heartbeats roughly every 41s,
// so this tolerates three missed beats before anything is declared unhealthy —
// long enough that an ordinary reconnect does not trip it.
const GatewayStaleAfter = 3 * time.Minute

// GatewayStater reports Discord gateway liveness. *discordgo.Session satisfies
// it through the adapter in this package; tests substitute their own so the
// dead-gateway states can be exercised without a real connection.
type GatewayStater interface {
	// LastHeartbeatAck is when Discord last acknowledged a heartbeat. It is
	// the zero time until the first ack arrives.
	LastHeartbeatAck() time.Time
	// HeartbeatLatency is the round trip of the last completed heartbeat.
	HeartbeatLatency() time.Duration
}

var (
	staterMu sync.RWMutex
	stater   GatewayStater

	// DiscordConnected tracks whether the gateway believes it is connected.
	// It is a coarse signal: a wedged connection can leave it at 1, which is
	// why health is decided by heartbeat staleness instead.
	DiscordConnected = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tbd_bot_discord_connected",
		Help: "1 when the Discord gateway reports itself connected, 0 otherwise.",
	})

	// GatewayEventsTotal counts gateway lifecycle transitions, which is how a
	// reconnect storm is told apart from a single clean outage after the fact.
	GatewayEventsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_discord_gateway_events_total",
		Help: "Total number of Discord gateway lifecycle events, by event.",
	}, []string{"event"})

	// ExternalAPIFailuresTotal records failed calls to services the bot does
	// not control. A token that stops working shows up here as a rate rather
	// than as silence.
	ExternalAPIFailuresTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_external_api_failures_total",
		Help: "Total number of failed external API calls, by api and reason.",
	}, []string{"api", "reason"})

	lastHeartbeatAckTimestamp = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "tbd_bot_discord_last_heartbeat_ack_timestamp_seconds",
		Help: "Unix timestamp of the last acknowledged Discord heartbeat, or 0 if none.",
	}, func() float64 {
		ack, ok := currentAck()
		if !ok || ack.IsZero() {
			return 0
		}
		return float64(ack.Unix())
	})

	heartbeatLatencySeconds = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "tbd_bot_discord_heartbeat_latency_seconds",
		Help: "Round trip of the last completed Discord heartbeat, in seconds.",
	}, func() float64 {
		staterMu.RLock()
		s := stater
		staterMu.RUnlock()
		if s == nil {
			return 0
		}
		return s.HeartbeatLatency().Seconds()
	})
)

// SetGatewayStater installs the source of gateway liveness. Passing nil clears
// it, which reads as unhealthy rather than as unknown.
func SetGatewayStater(g GatewayStater) {
	staterMu.Lock()
	defer staterMu.Unlock()
	stater = g
}

func currentAck() (time.Time, bool) {
	staterMu.RLock()
	s := stater
	staterMu.RUnlock()
	if s == nil {
		return time.Time{}, false
	}
	return s.LastHeartbeatAck(), true
}

// GatewayHealthy reports whether the Discord gateway is alive, and why not
// when it is not. Every unhealthy answer decays into existence on its own: no
// handler has to fire for the bot to be recognised as dead, which is the
// failure the /metrics-based healthcheck could not see.
func GatewayHealthy() (bool, string) {
	ack, ok := currentAck()
	if !ok {
		return false, "no Discord session registered"
	}
	if ack.IsZero() {
		return false, "no Discord heartbeat has ever been acknowledged"
	}
	if age := time.Since(ack); age > GatewayStaleAfter {
		return false, fmt.Sprintf("last Discord heartbeat acknowledged %s ago, over the %s limit", age.Truncate(time.Second), GatewayStaleAfter)
	}
	return true, ""
}

// SetDiscordConnected records the gateway's own view of its connection.
func SetDiscordConnected(connected bool) {
	if connected {
		DiscordConnected.Set(1)
		return
	}
	DiscordConnected.Set(0)
}

// IncGatewayEvent counts a gateway lifecycle event such as connect, disconnect
// or resumed.
func IncGatewayEvent(event string) {
	GatewayEventsTotal.WithLabelValues(event).Inc()
}

// IncExternalAPIFailure counts a failed call to an external API.
func IncExternalAPIFailure(api, reason string) {
	ExternalAPIFailuresTotal.WithLabelValues(api, reason).Inc()
}

// IncDiscordRESTFailure counts a failed Discord REST call, recovering the
// status code from discordgo's error type. Callers only ever hold the error,
// never the response.
func IncDiscordRESTFailure(err error) {
	var restErr *discordgo.RESTError
	if errors.As(err, &restErr) && restErr.Response != nil {
		IncExternalAPIFailure("discord", HTTPFailureReason(restErr.Response.StatusCode))
		return
	}
	IncExternalAPIFailure("discord", "transport_error")
}

// HTTPFailureReason collapses a status code into one of a fixed set of label
// values. The raw code would work as a label right up until an upstream
// started returning a wide spread of them and the series count followed.
func HTTPFailureReason(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusTooManyRequests:
		return "rate_limited"
	}
	switch {
	case status >= 500:
		return "server_error"
	case status >= 400:
		return "client_error"
	}
	return "unknown"
}
