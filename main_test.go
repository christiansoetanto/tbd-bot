package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
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

	if !strings.Contains(dockerfile, "http://localhost:8080/metrics") && !strings.Contains(dockerfile, ":8080/metrics") {
		t.Errorf("expected HEALTHCHECK to query http://localhost:8080/metrics")
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
}
