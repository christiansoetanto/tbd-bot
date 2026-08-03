package util

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// heartbeatTimeout bounds a single ping. The switch is a safety net, so a
// hung request must never hold the cron goroutine that drives it.
const heartbeatTimeout = 10 * time.Second

var (
	// ExternalHeartbeatTotal counts ping attempts by outcome. success and fail
	// both mean the ping was delivered; error means it was not, which is the
	// case that leaves the external switch blind.
	ExternalHeartbeatTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_external_heartbeat_total",
		Help: "Total external dead-man's-switch pings, by result.",
	}, []string{"result"})

	// externalHeartbeatLastPing is when a ping was last delivered, healthy or
	// not. Tracking only healthy pings would make a gateway outage read as a
	// broken pinger too, and send the same outage to Discord twice.
	externalHeartbeatLastPing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tbd_bot_external_heartbeat_last_ping_timestamp_seconds",
		Help: "Unix timestamp of the last delivered external heartbeat ping, or 0 if none.",
	})

	// externalHeartbeatEnabled is 0 when no ping URL is configured. Without it
	// an unset URL looks exactly like a healthy system: nothing errors, nothing
	// is delivered, and every dashboard stays green.
	externalHeartbeatEnabled = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tbd_bot_external_heartbeat_enabled",
		Help: "1 when an external heartbeat ping URL is configured, 0 otherwise.",
	})
)

// Heartbeat pings an external dead-man's switch so that the machine going dark
// is itself an alert. Everything else the bot reports runs on the same Mac
// Mini as Prometheus and Grafana, and so cannot report that Mac being dead.
type Heartbeat struct {
	successURL string
	failURL    string
	client     *http.Client
}

// NewHeartbeat builds a heartbeat for a healthchecks.io-style ping URL. An
// empty URL yields a disabled heartbeat that reports itself as disabled rather
// than silently doing nothing.
func NewHeartbeat(rawURL string) *Heartbeat {
	url := strings.TrimSpace(rawURL)
	url = strings.TrimRight(url, "/")
	if url == "" {
		externalHeartbeatEnabled.Set(0)
		return &Heartbeat{}
	}
	externalHeartbeatEnabled.Set(1)
	return &Heartbeat{
		successURL: url,
		failURL:    url + "/fail",
		client:     &http.Client{Timeout: heartbeatTimeout},
	}
}

// Enabled reports whether a ping URL was configured.
func (h *Heartbeat) Enabled() bool { return h.successURL != "" }

// Ping reports gateway liveness to the external switch. It stays silent until
// the gateway has connected at least once: GatewayHealthy is false for the
// ~26s a deploy needs to reach Ready, and pinging /fail there would alert on
// every deploy until the alerts stopped being read.
func (h *Heartbeat) Ping(ctx context.Context) {
	if !h.Enabled() || !GatewayEverConnected() {
		return
	}

	url, body, result := h.successURL, "", "success"
	if healthy, reason := GatewayHealthy(); !healthy {
		url, body, result = h.failURL, reason, "fail"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		ExternalHeartbeatTotal.WithLabelValues("error").Inc()
		return
	}
	res, err := h.client.Do(req)
	if err != nil {
		ExternalHeartbeatTotal.WithLabelValues("error").Inc()
		return
	}
	defer res.Body.Close()
	// Drain so the connection can be reused by the next minute's ping.
	_, _ = io.Copy(io.Discard, res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		// A wrong or deleted UUID answers 404. Counting that as delivered
		// would leave the one metric that proves the switch works lying.
		ExternalHeartbeatTotal.WithLabelValues("error").Inc()
		return
	}
	ExternalHeartbeatTotal.WithLabelValues(result).Inc()
	externalHeartbeatLastPing.SetToCurrentTime()
}
