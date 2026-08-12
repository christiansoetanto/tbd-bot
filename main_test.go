package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/christiansoetanto/tbd-bot/util"
)

type mockDiscordSession struct {
	closed bool
}

func (m *mockDiscordSession) Close() error {
	m.closed = true
	return nil
}

func TestMetricsEndpoint(t *testing.T) {
	handler := setupRoutes()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "go_goroutines") {
		t.Fatalf("expected metrics response to contain 'go_goroutines', got:\n%s", body)
	}
	if !strings.Contains(body, "tbd_bot_qa_moves_total") {
		t.Fatalf("expected metrics response to contain 'tbd_bot_qa_moves_total', got:\n%s", body)
	}
	if !strings.Contains(body, "tbd_bot_users_vetted_total") {
		t.Fatalf("expected metrics response to contain 'tbd_bot_users_vetted_total', got:\n%s", body)
	}
}

// staterStub drives gateway liveness from the test rather than from a real
// connection.
type staterStub struct{ ack time.Time }

func (s staterStub) LastHeartbeatAck() time.Time     { return s.ack }
func (s staterStub) HeartbeatLatency() time.Duration { return 0 }

// Docker called the container healthy for 33 hours while the bot was dead in
// Discord, because the probe only proved the HTTP server was up. /health has
// to fail when the gateway is stale or the healthcheck is decorative.
func TestHealthEndpoint_ReflectsGatewayLiveness(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })
	handler := setupRoutes()

	util.SetGatewayStater(staterStub{ack: time.Now()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected %d with a live gateway, got %d", http.StatusOK, rec.Code)
	}

	util.SetGatewayStater(staterStub{ack: time.Now().Add(-util.GatewayStaleAfter - time.Minute)})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected %d with a stale gateway, got %d", http.StatusServiceUnavailable, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "heartbeat") {
		t.Errorf("expected the body to explain the failure, got %q", rec.Body.String())
	}
}

// /metrics must stay unconditional: Prometheus has to keep scraping through an
// outage, otherwise the series needed to alert on it stops existing.
func TestMetricsEndpointServesWhileGatewayIsDown(t *testing.T) {
	t.Cleanup(func() { util.SetGatewayStater(nil) })
	util.SetGatewayStater(staterStub{ack: time.Now().Add(-util.GatewayStaleAfter - time.Minute)})

	rec := httptest.NewRecorder()
	setupRoutes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /metrics to stay available during an outage, got %d", rec.Code)
	}
}

// scripts/verify_dockerfile.sh asserts the same healthcheck contract as the
// Dockerfile, so the two drift apart silently — moving the probe to /health
// left the script failing on a correct Dockerfile. Same shape as the CI
// go-version drift: a second file encoding a contract nothing cross-checks.
func TestVerifyScriptAgreesWithDockerfileHealthcheck(t *testing.T) {
	script, err := os.ReadFile("scripts/verify_dockerfile.sh")
	if err != nil {
		t.Fatalf("failed to read verify_dockerfile.sh: %v", err)
	}
	s := string(script)
	if !strings.Contains(s, ":8080/health") {
		t.Error("verify_dockerfile.sh does not require the /health endpoint the Dockerfile probes")
	}
	// The stale assertion, verbatim: the script erroring because the
	// HEALTHCHECK does *not* target /metrics is the defect, not any mention
	// of the word.
	if strings.Contains(s, "HEALTHCHECK does not target http://localhost:8080/metrics") {
		t.Error("verify_dockerfile.sh still errors when the HEALTHCHECK does not target /metrics")
	}
}

func TestDockerfileHealthcheckProbesGatewayAwareEndpoint(t *testing.T) {
	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("failed to read Dockerfile: %v", err)
	}
	healthcheck := regexp.MustCompile(`(?ms)^HEALTHCHECK.*?\n\n`).FindString(string(content))
	if healthcheck == "" {
		t.Fatal("no HEALTHCHECK instruction found in Dockerfile")
	}
	if !strings.Contains(healthcheck, "/health") {
		t.Errorf("HEALTHCHECK must probe /health so a dead gateway fails it, got:\n%s", healthcheck)
	}
	if strings.Contains(healthcheck, "/metrics") {
		t.Errorf("HEALTHCHECK still probes /metrics, which answers 200 with the gateway dead:\n%s", healthcheck)
	}
}

func TestGracefulShutdown_SIGINT(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	srv := &http.Server{
		Handler: setupRoutes(),
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Serve(listener)
	}()

	mockSession := &mockDiscordSession{}
	sigChan := make(chan os.Signal, 1)
	sigChan <- syscall.SIGINT

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = gracefulShutdown(ctx, srv, mockSession, sigChan)
	if err != nil {
		t.Fatalf("expected gracefulShutdown to succeed, got error: %v", err)
	}

	if !mockSession.closed {
		t.Errorf("expected Discord session to be closed on SIGINT")
	}

	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("expected http.ErrServerClosed, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("expected HTTP server to exit after shutdown")
	}
}

func TestGracefulShutdown_SIGTERM(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	srv := &http.Server{
		Handler: setupRoutes(),
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Serve(listener)
	}()

	mockSession := &mockDiscordSession{}
	sigChan := make(chan os.Signal, 1)
	sigChan <- syscall.SIGTERM

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = gracefulShutdown(ctx, srv, mockSession, sigChan)
	if err != nil {
		t.Fatalf("expected gracefulShutdown to succeed, got error: %v", err)
	}

	if !mockSession.closed {
		t.Errorf("expected Discord session to be closed on SIGTERM")
	}

	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("expected http.ErrServerClosed, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("expected HTTP server to exit after shutdown")
	}
}

func TestDockerfile_MultiStageAndHealthcheck(t *testing.T) {
	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("failed to read Dockerfile: %v", err)
	}

	dockerfile := string(content)

	if !strings.Contains(dockerfile, "FROM golang:") {
		t.Errorf("expected Dockerfile to build from a golang base image")
	}

	if !strings.Contains(dockerfile, "CGO_ENABLED=0") {
		t.Errorf("expected Dockerfile to set CGO_ENABLED=0")
	}

	if !strings.Contains(dockerfile, "HEALTHCHECK") {
		t.Errorf("expected Dockerfile to contain HEALTHCHECK directive")
	}

	// /metrics is served only after dbot.Init completes its Discord round trip,
	// so too short a start period reports unhealthy during a healthy boot.
	startPeriod := regexp.MustCompile(`--start-period=(\d+)s`).FindStringSubmatch(dockerfile)
	if startPeriod == nil {
		t.Errorf("expected HEALTHCHECK to set --start-period")
	} else if secs, err := strconv.Atoi(startPeriod[1]); err != nil {
		t.Errorf("malformed --start-period value %q", startPeriod[1])
	} else if secs < 30 {
		t.Errorf("HEALTHCHECK --start-period=%ds is too short to cover Discord command registration", secs)
	}

	if !strings.Contains(dockerfile, "wget") {
		t.Errorf("expected HEALTHCHECK to use wget")
	}

	// Superseded on 2026-08-03: this asserted /metrics, which answers 200 off
	// the HTTP server alone and so passed for the whole 33-hour outage.
	// TestDockerfileHealthcheckProbesGatewayAwareEndpoint owns the endpoint
	// choice now; this only checks the port is still right.
	if !strings.Contains(dockerfile, "http://localhost:8080/health") && !strings.Contains(dockerfile, ":8080/health") {
		t.Errorf("expected HEALTHCHECK to query http://localhost:8080/health")
	}
}

// parseMajorMinor extracts the leading "X.Y" of a Go version string.
func parseMajorMinor(t *testing.T, version string) (int, int) {
	t.Helper()
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		t.Fatalf("malformed go version %q", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		t.Fatalf("malformed major version in %q: %v", version, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		t.Fatalf("malformed minor version in %q: %v", version, err)
	}
	return major, minor
}

// The Docker build fails outright when the builder image is older than the
// version required by go.mod, so the two must be checked against each other
// rather than pinned to a literal tag.
func TestDockerfileGoVersionSatisfiesGoMod(t *testing.T) {
	goMod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	modMatch := regexp.MustCompile(`(?m)^go\s+(\d+\.\d+(?:\.\d+)?)`).FindStringSubmatch(string(goMod))
	if modMatch == nil {
		t.Fatalf("no go directive found in go.mod")
	}

	dockerfile, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("failed to read Dockerfile: %v", err)
	}
	imgMatch := regexp.MustCompile(`FROM\s+golang:(\d+\.\d+(?:\.\d+)?)`).FindStringSubmatch(string(dockerfile))
	if imgMatch == nil {
		t.Fatalf("no versioned golang base image found in Dockerfile")
	}

	modMajor, modMinor := parseMajorMinor(t, modMatch[1])
	imgMajor, imgMinor := parseMajorMinor(t, imgMatch[1])

	if imgMajor < modMajor || (imgMajor == modMajor && imgMinor < modMinor) {
		t.Errorf("Dockerfile builder golang:%s is older than go.mod requirement go %s", imgMatch[1], modMatch[1])
	}
}

// CI installs its own toolchain, so it drifts from go.mod independently of the
// Dockerfile. A toolchain older than the go directive fails at module parse
// time, before any build error the developer would recognize.
func TestCIWorkflowGoVersionSatisfiesGoMod(t *testing.T) {
	goMod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	modMatch := regexp.MustCompile(`(?m)^go\s+(\d+\.\d+(?:\.\d+)?)`).FindStringSubmatch(string(goMod))
	if modMatch == nil {
		t.Fatalf("no go directive found in go.mod")
	}

	workflowPath := ".github/workflows/go.yml"
	workflow, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", workflowPath, err)
	}
	verMatch := regexp.MustCompile(`go-version:\s*['"]?(\d+\.\d+(?:\.\d+)?)`).FindStringSubmatch(string(workflow))
	if verMatch == nil {
		t.Fatalf("no go-version found in %s", workflowPath)
	}

	modMajor, modMinor := parseMajorMinor(t, modMatch[1])
	ciMajor, ciMinor := parseMajorMinor(t, verMatch[1])

	if ciMajor < modMajor || (ciMajor == modMajor && ciMinor < modMinor) {
		t.Errorf("%s go-version %s is older than go.mod requirement go %s", workflowPath, verMatch[1], modMatch[1])
	}
}

// grpcSecurityFloor is the lowest grpc-go release that carries no open
// advisory against this module. Three Dependabot alerts stacked up on
// google.golang.org/grpc while it sat at v1.53.0 — an authorization bypass via
// a missing leading slash in :path (critical, fixed 1.79.3), xDS RBAC and
// HTTP/2 issues (high, fixed 1.82.1), and HTTP/2 Rapid Reset (high, fixed
// 1.56.3). The highest of those is the floor.
const grpcSecurityFloor = "1.82.1"

// grpc is an indirect dependency, so nothing in this repo imports it and no
// build failure would ever announce a regression. Minimal version selection
// will happily walk it back down if some parent module is bumped to one that
// requires an older grpc, and the only symptom would be the alerts silently
// reopening. This asserts the version actually selected rather than the text
// in go.mod, because a require line can be present and not be what is built.
func TestGrpcVersionClearsSecurityFloor(t *testing.T) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Version}}", "google.golang.org/grpc").Output()
	if err != nil {
		t.Fatalf("failed to resolve selected grpc version: %v", err)
	}
	selected := strings.TrimSpace(string(out))

	if compareSemver(t, selected, grpcSecurityFloor) < 0 {
		t.Errorf("google.golang.org/grpc resolves to %s, below the security floor v%s", selected, grpcSecurityFloor)
	}
}

// compareSemver orders two versions numerically. Comparing them as strings
// looks like it works and does not: "v1.9.0" sorts above "v1.82.1".
func compareSemver(t *testing.T, a, b string) int {
	t.Helper()
	parse := func(v string) [3]int {
		t.Helper()
		match := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`).FindStringSubmatch(v)
		if match == nil {
			t.Fatalf("unparseable version %q", v)
		}
		var parts [3]int
		for i := 0; i < 3; i++ {
			n, err := strconv.Atoi(match[i+1])
			if err != nil {
				t.Fatalf("unparseable version %q: %v", v, err)
			}
			parts[i] = n
		}
		return parts
	}
	x, y := parse(a), parse(b)
	for i := 0; i < 3; i++ {
		if x[i] != y[i] {
			if x[i] < y[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func TestDockerComposeAndMonitoringSetup(t *testing.T) {
	composeBytes, err := os.ReadFile("docker-compose.yml")
	if err != nil {
		t.Fatalf("failed to read docker-compose.yml: %v", err)
	}
	compose := string(composeBytes)

	requiredComposeSubstrings := []string{
		"tbd-bot",
		"prometheus",
		"grafana",
		"GOMEMLIMIT=",
		"127.0.0.1:9400:3000",
		"--storage.tsdb.retention.time=14d",
		"--storage.tsdb.retention.size=1GB",
		"prometheus_data:",
		"grafana_data:",
	}
	for _, sub := range requiredComposeSubstrings {
		if !strings.Contains(compose, sub) {
			t.Errorf("docker-compose.yml missing expected substring: %q", sub)
		}
	}

	promBytes, err := os.ReadFile("prometheus/prometheus.yml")
	if err != nil {
		t.Fatalf("failed to read prometheus/prometheus.yml: %v", err)
	}
	prom := string(promBytes)
	if !strings.Contains(prom, "tbd-bot:8080") {
		t.Errorf("prometheus.yml missing target tbd-bot:8080")
	}

	dsBytes, err := os.ReadFile("grafana/provisioning/datasources/datasource.yml")
	if err != nil {
		t.Fatalf("failed to read grafana/provisioning/datasources/datasource.yml: %v", err)
	}
	if !strings.Contains(string(dsBytes), "http://prometheus:9090") {
		t.Errorf("datasource.yml missing http://prometheus:9090")
	}

	dbProvBytes, err := os.ReadFile("grafana/provisioning/dashboards/dashboard-provider.yml")
	if err != nil {
		t.Fatalf("failed to read grafana/provisioning/dashboards/dashboard-provider.yml: %v", err)
	}
	if !strings.Contains(string(dbProvBytes), "/etc/grafana/provisioning/dashboards") {
		t.Errorf("dashboard-provider.yml missing dashboard path")
	}

	envBytes, err := os.ReadFile(".env.example")
	if err != nil {
		t.Fatalf("failed to read .env.example: %v", err)
	}
	envEx := string(envBytes)
	if !strings.Contains(envEx, "GOMEMLIMIT=") || !strings.Contains(envEx, "PORT=") {
		t.Errorf(".env.example missing expected variables")
	}
}

// The dashboard is what gets looked at during an incident. If the gateway and
// external-API series are not on it, the next outage looks exactly like the
// last one: every panel healthy, bot dead.
func TestDashboardCoversGatewayAndAPIFailures(t *testing.T) {
	raw, err := os.ReadFile("grafana/provisioning/dashboards/bot-dashboard.json")
	if err != nil {
		t.Fatalf("failed to read dashboard: %v", err)
	}
	dashboard := string(raw)

	requiredExprSubstrings := []string{
		"tbd_bot_discord_last_heartbeat_ack_timestamp_seconds",
		"tbd_bot_discord_connected",
		"tbd_bot_discord_heartbeat_latency_seconds",
		"tbd_bot_discord_gateway_events_total",
		"tbd_bot_external_api_failures_total",
	}
	for _, sub := range requiredExprSubstrings {
		if !strings.Contains(dashboard, sub) {
			t.Errorf("dashboard has no panel querying %s", sub)
		}
	}

	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("dashboard is not valid JSON: %v", err)
	}
}

// Alerting is the only part of this that reaches the user when they are not
// looking at Grafana, so its provisioning has to be checked into the repo
// rather than clicked together in the UI where a volume wipe loses it.
func TestAlertingProvisioned(t *testing.T) {
	contactPoints, err := os.ReadFile("grafana/provisioning/alerting/contact-points.yml")
	if err != nil {
		t.Fatalf("failed to read contact-points.yml: %v", err)
	}
	cp := string(contactPoints)
	if !strings.Contains(cp, "GRAFANA_DISCORD_WEBHOOK_URL") {
		t.Error("contact point must read the webhook from GRAFANA_DISCORD_WEBHOOK_URL, not hardcode it")
	}
	if strings.Contains(cp, "discord.com/api/webhooks/") {
		t.Error("contact-points.yml contains a literal webhook URL; that is a credential and must stay in .env")
	}

	rules, err := os.ReadFile("grafana/provisioning/alerting/rules.yml")
	if err != nil {
		t.Fatalf("failed to read rules.yml: %v", err)
	}
	r := string(rules)
	requiredRuleSubstrings := []string{
		"tbd_bot_discord_last_heartbeat_ack_timestamp_seconds",
		"tbd_bot_external_api_failures_total",
		`up{job="tbd-bot"}`,
	}
	for _, sub := range requiredRuleSubstrings {
		if !strings.Contains(r, sub) {
			t.Errorf("rules.yml missing expected content: %q", sub)
		}
	}

	// Every signal that failed on 08-01 failed by going quiet. A rule that
	// treats missing data as "fine" reproduces exactly that, so the
	// availability rules have to alert on absence.
	if !strings.Contains(r, "noDataState: Alerting") {
		t.Error("rules.yml has no rule that alerts on missing data")
	}
	if strings.Count(r, "noDataState:") < 3 {
		t.Errorf("expected every rule to declare noDataState explicitly, found %d", strings.Count(r, "noDataState:"))
	}

	if _, err := os.ReadFile("grafana/provisioning/alerting/notification-policies.yml"); err != nil {
		t.Fatalf("failed to read notification-policies.yml: %v", err)
	}

	// A rule that fires into a contact point Grafana never received is the
	// same silence as no rule at all.
	compose, err := os.ReadFile("docker-compose.yml")
	if err != nil {
		t.Fatalf("failed to read docker-compose.yml: %v", err)
	}
	if !strings.Contains(string(compose), "GRAFANA_DISCORD_WEBHOOK_URL") {
		t.Error("docker-compose.yml does not pass GRAFANA_DISCORD_WEBHOOK_URL to Grafana")
	}

	envExample, err := os.ReadFile(".env.example")
	if err != nil {
		t.Fatalf("failed to read .env.example: %v", err)
	}
	if !strings.Contains(string(envExample), "GRAFANA_DISCORD_WEBHOOK_URL") {
		t.Error(".env.example does not document GRAFANA_DISCORD_WEBHOOK_URL")
	}
}

func TestDocumentation_GrafanaAndRunnerSecurity(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	readme := string(readmeBytes)

	requiredSubstrings := []string{
		"127.0.0.1:9400",
		"Grafana",
		"Fork",
		"runner",
	}

	for _, sub := range requiredSubstrings {
		if !strings.Contains(readme, sub) {
			t.Errorf("README.md missing expected documentation section containing: %q", sub)
		}
	}
}

func TestGrafanaDashboardJSON(t *testing.T) {
	dashPath := "grafana/provisioning/dashboards/bot-dashboard.json"
	data, err := os.ReadFile(dashPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", dashPath, err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("bot-dashboard.json is not valid JSON: %v", err)
	}

	raw := string(data)
	requiredMetrics := []string{
		"tbd_bot_handler_requests_total",
		"tbd_bot_handler_duration_seconds",
		"tbd_bot_qa_moves_total",
		"tbd_bot_users_vetted_total",
		"tbd_bot_ts_users_vetted_total",
		"tbd_bot_cron_executions_total",
		"tbd_bot_cm_actions_total",
		"tbd_bot_messages_processed_total",
		"tbd_bot_component_interactions_total",
		// Runtime panels. These are always populated, so the dashboard shows
		// something real even before any bot event has occurred.
		"go_goroutines",
		"process_resident_memory_bytes",
		"process_start_time_seconds",
	}

	for _, metric := range requiredMetrics {
		if !strings.Contains(raw, metric) {
			t.Errorf("bot-dashboard.json missing Prometheus metric query for: %q", metric)
		}
	}

	// Labelled counters have no series at all until their first increment, so a
	// panel querying one renders "No data" — indistinguishable from a broken
	// dashboard. Each must fall back to an explicit zero.
	needsZeroFallback := []string{
		"tbd_bot_cron_executions_total",
		"tbd_bot_cm_actions_total",
		"tbd_bot_messages_processed_total",
		"tbd_bot_component_interactions_total",
	}

	panels, _ := parsed["panels"].([]any)
	for _, metric := range needsZeroFallback {
		found := false
		for _, p := range panels {
			panel, ok := p.(map[string]any)
			if !ok {
				continue
			}
			targets, _ := panel["targets"].([]any)
			for _, tg := range targets {
				target, ok := tg.(map[string]any)
				if !ok {
					continue
				}
				expr, _ := target["expr"].(string)
				if strings.Contains(expr, metric) {
					found = true
					if !strings.Contains(expr, "or vector(0)") {
						t.Errorf("panel querying %q must fall back to `or vector(0)`, got: %s", metric, expr)
					}
				}
			}
		}
		if !found {
			t.Errorf("no panel queries %q", metric)
		}
	}
}

func TestDeployWorkflow(t *testing.T) {
	deployPath := ".github/workflows/deploy.yml"
	data, err := os.ReadFile(deployPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", deployPath, err)
	}

	content := string(data)

	requiredSubstrings := []string{
		"master",
		"self-hosted",
		"cancel-in-progress: true",
		// Compose V2 plugin form. Docker Desktop no longer ships the standalone
		// docker-compose binary, and the runner has no shell alias for it.
		"docker compose up -d --build",
		"docker inspect --format='{{json .State.Health.Status}}'",
		"\"healthy\"",
		"~/tbd-bot-secrets/.env",
	}

	for _, sub := range requiredSubstrings {
		if !strings.Contains(content, sub) {
			t.Errorf("deploy.yml missing required constraint substring: %q", sub)
		}
	}

	// The compose command must stay service-less. Narrowing it to
	// "... --build tbd-bot" still deploys the bot, so nothing looks broken,
	// but Grafana never gets recreated and the alerting provisioning in
	// grafana/provisioning/alerting/ only ever exists on whichever machine
	// last ran compose by hand.
	upCmd := regexp.MustCompile(`docker compose up -d --build(.*)`).FindStringSubmatch(content)
	if upCmd == nil {
		t.Fatal("deploy.yml has no `docker compose up -d --build` command")
	}
	if strings.TrimSpace(upCmd[1]) != "" {
		t.Errorf("deploy.yml limits the deploy to %q; it must recreate the whole stack so Grafana picks up alerting provisioning", strings.TrimSpace(upCmd[1]))
	}
}

// The external switch is the only watcher that outlives this Mac, so the two
// ways it can be silently absent — no URL, or a URL that stops working — both
// need a rule of their own.
func TestExternalHeartbeatProvisioned(t *testing.T) {
	rules, err := os.ReadFile("grafana/provisioning/alerting/rules.yml")
	if err != nil {
		t.Fatalf("failed to read rules.yml: %v", err)
	}
	r := string(rules)
	for _, sub := range []string{
		"tbd_bot_external_heartbeat_enabled",
		"tbd_bot_external_heartbeat_last_ping_timestamp_seconds",
		"tbd-bot-heartbeat-disabled",
		"tbd-bot-heartbeat-not-delivering",
	} {
		if !strings.Contains(r, sub) {
			t.Errorf("rules.yml missing expected content: %q", sub)
		}
	}

	// tbd-bot-target-down already owns "the bot is absent". With the policy
	// grouping by alertname, an Alerting no-data state on these two would turn
	// a single outage into three separate Discord messages.
	for _, uid := range []string{"tbd-bot-heartbeat-disabled", "tbd-bot-heartbeat-not-delivering"} {
		idx := strings.Index(r, "uid: "+uid)
		if idx < 0 {
			t.Fatalf("rule %s not found", uid)
		}
		rest := r[idx:]
		if end := strings.Index(rest, "\n      - uid:"); end > 0 {
			rest = rest[:end]
		}
		if !strings.Contains(rest, "noDataState: OK") {
			t.Errorf("rule %s must use noDataState: OK so an outage is not reported twice", uid)
		}
	}

	envExample, err := os.ReadFile(".env.example")
	if err != nil {
		t.Fatalf("failed to read .env.example: %v", err)
	}
	env := string(envExample)
	if !strings.Contains(env, "HEALTHCHECKS_PING_URL") {
		t.Error(".env.example does not document HEALTHCHECKS_PING_URL")
	}

	// The bot reads the ping URL through env_file, so compose needs no change
	// — but only as long as env_file stays.
	compose, err := os.ReadFile("docker-compose.yml")
	if err != nil {
		t.Fatalf("failed to read docker-compose.yml: %v", err)
	}
	if !strings.Contains(string(compose), "env_file") {
		t.Error("docker-compose.yml no longer passes .env to the bot, so HEALTHCHECKS_PING_URL cannot reach it")
	}
}

// A ping URL in a tracked file is a credential leak in the direction that
// matters least visibly: it would let anyone suppress the alert.
func TestNoPingURLInTrackedFiles(t *testing.T) {
	out, err := exec.Command("git", "grep", "-lE", `hc-ping\.com/[0-9a-f]{8}`).CombinedOutput()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		t.Errorf("ping URLs found in tracked files:\n%s", out)
	}
}

func TestDashboardCoversExternalHeartbeat(t *testing.T) {
	content, err := os.ReadFile("grafana/provisioning/dashboards/bot-dashboard.json")
	if err != nil {
		t.Fatalf("failed to read the dashboard: %v", err)
	}
	var dashboard map[string]any
	if err := json.Unmarshal(content, &dashboard); err != nil {
		t.Fatalf("dashboard is not valid JSON: %v", err)
	}
	body := string(content)
	for _, metric := range []string{
		"tbd_bot_external_heartbeat_enabled",
		"tbd_bot_external_heartbeat_last_ping_timestamp_seconds",
		"tbd_bot_external_heartbeat_total",
	} {
		if !strings.Contains(body, metric) {
			t.Errorf("dashboard has no panel for %s", metric)
		}
	}
}

// `docker compose up -d --build` only recreates containers whose image or
// config changed. A change under grafana/provisioning/ changes neither, so
// Grafana keeps running with the rules it read at startup — which is how the
// two heartbeat rules deployed on 08-03 and were simply absent afterwards.
// TestDeployWorkflow cannot catch this: the compose command is correct.
func TestDeployWorkflowReloadsGrafanaProvisioning(t *testing.T) {
	content, err := os.ReadFile(".github/workflows/deploy.yml")
	if err != nil {
		t.Fatalf("failed to read deploy.yml: %v", err)
	}
	workflow := string(content)

	if !strings.Contains(workflow, "docker compose restart grafana") {
		t.Error("deploy.yml never reloads Grafana, so alerting and dashboard changes deploy to disk and are never read")
	}

	// Order matters: reloading before the new files are in place would read
	// the old ones and look like it worked.
	up := strings.Index(workflow, "docker compose up -d --build")
	reload := strings.Index(workflow, "docker compose restart grafana")
	if up < 0 || reload < 0 {
		t.Fatal("deploy.yml is missing the compose up or the Grafana reload")
	}
	if reload < up {
		t.Error("deploy.yml reloads Grafana before composing the stack; it must come after")
	}

	// Restarting the whole stack instead would take the bot down on every
	// deploy to pick up a Grafana-only change.
	if strings.Contains(workflow, "docker compose restart\n") || strings.Contains(workflow, "docker compose restart tbd-bot") {
		t.Error("deploy.yml restarts more than Grafana; a provisioning change must not cost bot downtime")
	}
}

// dockerignoreMatches reimplements the directory-prefix semantics dockerd
// uses for plain (non-**) .dockerignore patterns: a pattern excludes a path
// if its components glob-match a PREFIX of the path's components, so "docs"
// (one component) excludes the whole subtree, while "*.md" (also one
// component) only ever matches root-level files, never nested ones.
func dockerignoreMatches(pattern, path string) bool {
	patParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(path, "/")
	if len(patParts) > len(pathParts) {
		return false
	}
	for i, p := range patParts {
		if ok, err := filepath.Match(p, pathParts[i]); err != nil || !ok {
			return false
		}
	}
	return true
}

func anyDockerignorePatternMatches(patterns []string, path string) bool {
	for _, pattern := range patterns {
		if dockerignoreMatches(pattern, path) {
			return true
		}
	}
	return false
}

// deploy.yml runs `docker compose up -d --build` on every push to master
// with no path filter, and Dockerfile:14 is `COPY . .`, so anything left in
// the build context that changes invalidates that layer and restarts the
// live bot. This test pins both directions of .dockerignore: docs and
// markdown must stay excluded (the regression this guards), and no pattern
// may reach a file the Go build actually needs, since that would silently
// ship a broken image.
func TestDockerignoreExcludesDocsButKeepsGoBuildInputs(t *testing.T) {
	ignoreBytes, err := os.ReadFile(".dockerignore")
	if err != nil {
		t.Fatalf("failed to read .dockerignore: %v", err)
	}
	var patterns []string
	for _, line := range strings.Split(string(ignoreBytes), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}

	out, err := exec.Command("git", "ls-files").Output()
	if err != nil {
		t.Fatalf("git ls-files failed: %v", err)
	}
	tracked := strings.Fields(string(out))

	for _, f := range tracked {
		isNonBuildInput := strings.HasSuffix(f, ".md") ||
			strings.HasPrefix(f, "docs/") ||
			strings.HasPrefix(f, "grafana/") ||
			strings.HasPrefix(f, "prometheus/")
		if !isNonBuildInput {
			continue
		}
		if !anyDockerignorePatternMatches(patterns, f) {
			t.Errorf(".dockerignore does not exclude %q; a commit touching only docs/grafana/prometheus would still invalidate the build context and restart the bot", f)
		}
	}

	for _, f := range tracked {
		isBuildInput := f == "go.mod" || f == "go.sum" || strings.HasSuffix(f, ".go")
		if !isBuildInput {
			continue
		}
		if anyDockerignorePatternMatches(patterns, f) {
			t.Errorf(".dockerignore excludes %q, which `go build -o tbd-bot .` needs; the image would fail to build on deploy", f)
		}
	}
}
