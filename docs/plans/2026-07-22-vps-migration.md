# Plan: Azure App Service to Mac Mini Migration

## Goal
Migrate the `tbd-bot` deployment architecture from the Azure Web App `tbdbot-cicd` to an M4 Mac Mini home server. This involves optimizing the Docker container, establishing deep application metrics (Prometheus/Grafana), and setting up automated CI/CD via a local GitHub Runner.

## Architecture
- **Metrics:** `prometheus/client_golang` integrated into the existing port 8080 HTTP server. Includes business logic counters and RED decorators for Discord events.
- **Docker Compose:** Standard orchestrator (`docker-compose.yml`) mapping to a `.env` file for secrets. Runs the bot (with `GOMEMLIMIT` and memory limits), Prometheus (14d/1GB limits), and Grafana (IaC Provisioning on `127.0.0.1:3000`).
- **CI/CD:** GitHub Actions targeting a `self-hosted` runner running on the Mac Mini. Includes concurrency control, secure `.env` injection, and true Docker health-check polling.

## Global Constraints
- The Mac Mini must maintain persistent internet access for the Discord WebSocket session.
- Secrets (`BOTTOKEN`, etc.) must be stored securely in a local `.env` file on the Mac Mini, and securely injected into the GitHub Actions workspace.
- Deployments must be automated via GitHub Actions. **Self-hosted runner security must be enforced** by strictly tying workflows to the `master` branch and disabling workflows from fork Pull Requests via GitHub Repository Settings.
- The system and application timezone must remain in UTC.
- The bot must avoid Discord Gateway bans by implementing a graceful `SIGTERM` shutdown that calls `discordgo.Session.Close()`.
- Only one instance may run against the production `BOTTOKEN` at a time. The Azure Web App must be stopped before the Mac Mini stack starts with production credentials.

## Progress

| Task | Wave | Status | Evidence |
|------|------|--------|----------|
| 1.1  | 1    | completed | `TestMetricsEndpoint` passed in `main_test.go` |
| 1.2  | 1    | completed | `TestBusinessMetricsRegistered` and `TestMetricsEndpoint` passed |
| 1.3  | 1    | completed | `TestREDMetricsDecorators` and `TestAllExpandedBusinessMetrics` passed in `util/metrics_test.go` |
| 1.4  | 1    | completed | `TestGracefulShutdown_SIGINT` and `TestGracefulShutdown_SIGTERM` passed in `main_test.go` |
| 2.1  | 2    | completed | `TestDockerfile_MultiStageAndHealthcheck` passed & `scripts/verify_dockerfile.sh` verified static rules + Go static build |
| 2.2  | 2    | completed | `TestDockerComposeAndMonitoringSetup` passed in `main_test.go` & valid `docker-compose.yml`, `prometheus/`, `grafana/` configs created |
| 2.3  | 2    | completed | `TestDocumentation_GrafanaAndRunnerSecurity` passed in `main_test.go` & `README.md` updated |
| 2.4  | 2    | completed | `TestGrafanaDashboardJSON` passed in `main_test.go` & valid JSON at `grafana/provisioning/dashboards/bot-dashboard.json` |
| 3.1  | 3    | completed | Verified `deploy.yml` with `actionlint` and `TestDeployWorkflow` in `main_test.go` |
| 4.1  | 4    | completed | `TestDockerfileGoVersionSatisfiesGoMod` failed on `golang:1.22` vs `go 1.25.0`, passes after the bump; `scripts/verify_dockerfile.sh` now compares the two versions instead of matching a literal tag |
| 4.2  | 4    | completed | `docs/mac-mini-setup.md` section 3 now enumerates all 7 variables read by the code plus `GF_SECURITY_ADMIN_PASSWORD`, matching `.env.example` |
| 4.3  | 4    | not started | Blocked on Docker being installed on the Mac Mini |
| 4.4  | 4    | not started | Blocked on 4.3 |
| 4.5  | 4    | not started | Blocked on 4.3; runner registration needs a token from the GitHub UI |
| 4.6  | 4    | not started | Blocked on 4.4 and 4.5 |

## Diagram

```mermaid
flowchart TD
  subgraph Wave 1
    1.1["1.1: Prometheus HTTP Handler"]
    1.2["1.2: Business Metrics Counters"]
    1.3["1.3: RED Metrics Decorators"]
    1.4["1.4: Graceful Shutdown"]
  end
  subgraph Wave 2
    2.1["2.1: Multi-stage Dockerfile"]
    2.2["2.2: Docker Compose & Monitoring Setup"]
    2.3["2.3: Documentation"]
    2.4["2.4: Baseline Dashboard JSON"]
  end
  subgraph Wave 3
    3.1["3.1: Self-Hosted Runner Deploy Action"]
  end
  subgraph Wave 4
    4.1["4.1: Builder Image Tracks go.mod"]
    4.2["4.2: Complete Secrets Documentation"]
    4.3["4.3: Install Docker"]
    4.4["4.4: Dry Run the Stack"]
    4.5["4.5: Register the Runner"]
    4.6["4.6: Cut Over From Azure"]
  end
  1.1 --> 1.2
  1.1 --> 1.3
  1.1 --> 2.1
  1.2 --> 2.2
  1.3 --> 2.2
  1.4 --> 2.1
  2.1 --> 2.2
  2.2 --> 2.3
  2.2 --> 2.4
  2.3 --> 3.1
  2.4 --> 3.1
  2.1 --> 4.1
  4.1 --> 4.4
  4.2 --> 4.4
  4.3 --> 4.4
  4.3 --> 4.5
  3.1 --> 4.5
  4.4 --> 4.6
  4.5 --> 4.6
```

## Wave 1

### Task 1.1: Prometheus HTTP Handler
**Files:** Modify `main.go`, `go.mod`, `go.sum`
**Write-scope:** `main.go`, `go.mod`, `go.sum`
**Consumes:** Existing port 8080 HTTP server setup
**Produces:** A `/metrics` endpoint serving default Go metrics using `github.com/prometheus/client_golang/prometheus/promhttp`.
**Seams:** HTTP boundary.
**Tests:** `go test` validating the metrics endpoint returns 200 MUST be written and fail first.
**Model tier:** flash

### Task 1.2: Business Metrics Counters
**Files:** Create `util/metrics.go`, modify `dbot/handler/sd_command.go`, `dbot/handler/reaction_add.go`, `dbot/init.go`
**Write-scope:** `util/metrics.go`, `dbot/handler/sd_command.go`, `dbot/handler/reaction_add.go`, `dbot/init.go`
**Consumes:** Task 1.1 Prometheus framework
**Produces:** Registered custom counters (`tbd_bot_qa_moves_total`, `tbd_bot_users_vetted_total`, etc.) that increment during their respective usecases.
**Seams:** Internal Go package boundary.
**Tests:** Verify counters increment logic via unit tests (MUST fail first).
**Model tier:** flash

### Task 1.3: RED Metrics Decorators
**Files:** Modify `util/metrics.go`, `dbot/init.go`, `dbot/handler/*.go`
**Write-scope:** `util/metrics.go`, `dbot/init.go`, `dbot/handler/`
**Consumes:** Task 1.1 Prometheus framework
**Produces:** Manual RED timers injected into handler registrations to track Latency (Histograms) and Success Rate/QPS (Counters) without slow reflection.
**Seams:** Metric injection boundary.
**Tests:** Unit test the decorators record metrics successfully (MUST fail first).
**Model tier:** flash

### Task 1.4: Graceful Shutdown
**Files:** Modify `main.go`
**Write-scope:** `main.go`
**Consumes:** Discordgo session initialization.
**Produces:** OS signal trapping for `SIGINT` and `SIGTERM` that explicitly calls `discordgo.Session.Close()` before exiting, preventing Gateway ban loops. HTTP server and Discord session each get independent 5-second timeout contexts to prevent one from starving the other. Discord `Close()` runs in a goroutine with a `select` to enforce the timeout.
**Seams:** OS process lifecycle.
**Tests:** Trigger a manual SIGINT in tests and verify `Close()` is called (MUST fail first).
**Model tier:** flash

## Wave 2

### Task 2.1: Multi-stage Dockerfile
**Files:** Modify `Dockerfile`
**Write-scope:** `Dockerfile`
**Consumes:** Go source code from Wave 1
**Produces:** A multi-stage Dockerfile compiling a static Go binary using an alpine `golang` builder image whose version is at least the `go` directive in `go.mod`, with `CGO_ENABLED=0`. Includes a native Docker `HEALTHCHECK` using `wget` against `:8080/metrics`. Runs as `USER nobody:nobody` (non-root). Sets `ENV TZ=UTC`.
**Seams:** Build boundary — must compile successfully.
**Tests:** Docker build execution and running `docker inspect` to verify health check syntax.
**Model tier:** flash

### Task 2.2: Docker Compose & Monitoring Setup
**Files:** Modify `docker-compose.yml`, create `prometheus/prometheus.yml`, `grafana/provisioning/datasources/datasource.yml`, `grafana/provisioning/dashboards/dashboard-provider.yml`, `.env.example`
**Write-scope:** `docker-compose.yml`, `prometheus/`, `grafana/`, `.env.example`
**Consumes:** Task 2.1 Image, Task 1.3 `/metrics` endpoint.
**Produces:** Orchestrated stack for bot (with memory limits and `GOMEMLIMIT`), prometheus (with named volumes and 1GB/14d limits), and grafana (with named volumes, IaC provisioning, bound to `127.0.0.1:3000`). Grafana datasource explicitly sets `uid: prometheus` to match dashboard JSON references. All ports bound to `127.0.0.1`.
**Seams:** Orchestration boundary.
**Tests:** Run `docker-compose config` to validate YAML before execution.
**Model tier:** flash

### Task 2.3: Documentation
**Files:** Modify `README.md`
**Write-scope:** `README.md`
**Consumes:** Task 2.2 Architecture
**Produces:** Clear instructions on how to access the Grafana dashboard locally on the Mac Mini network (`127.0.0.1`), and instructions to disable Fork PRs in GitHub repo settings.
**Seams:** Documentation boundary.
**Tests:** Manual proofread.
**Model tier:** flash

### Task 2.4: Baseline Dashboard JSON
**Files:** Create `grafana/provisioning/dashboards/bot-dashboard.json`
**Write-scope:** `grafana/provisioning/dashboards/bot-dashboard.json`
**Consumes:** Task 2.2 Provisioning setup.
**Produces:** A baseline JSON file that auto-loads in Grafana to visualize RED metrics.
**Seams:** Dashboard JSON validation.
**Tests:** Valid JSON syntax check.
**Model tier:** flash

## Wave 3

### Task 3.1: Self-Hosted Runner Deploy Action
**Files:** Modify `.github/workflows/deploy.yml`
**Write-scope:** `.github/workflows/deploy.yml`
**Consumes:** Docker Compose infrastructure from Task 2.2
**Produces:** A GitHub Action YAML that triggers on push to `master` targeting a `self-hosted` runner, executing `docker-compose up -d --build`. Health check timeout set to 120s to account for cold-boot Prometheus/Grafana settling.
**Seams:** Pipeline boundary.
**Tests:** Run `actionlint .github/workflows/deploy.yml` locally to validate syntax (MUST pass validation).
**Model tier:** flash
- [x] Add `concurrency: cancel-in-progress: true`.
- [x] Inject the `.env` file via copying from a secure external directory outside the GitHub workspace.
- [x] Add local `docker-compose up -d --build` step.
- [x] Add post-deployment health check polling `docker inspect --format='{{json .State.Health.Status}}'` until it returns `"healthy"`.

## Wave 4

Waves 1-3 produced files. Nothing in them ever executed against Docker, because Docker
was not installed on the Mac Mini — the Docker-related tests assert over file text
only. Wave 4 is the part that runs.

### Task 4.1: Builder Image Tracks go.mod
**Files:** Modify `Dockerfile`, `main_test.go`, `scripts/verify_dockerfile.sh`
**Write-scope:** `Dockerfile`, `main_test.go`, `scripts/verify_dockerfile.sh`
**Consumes:** Task 2.1
**Produces:** A builder image at least as new as the `go` directive in `go.mod`. The literal-tag assertion is replaced by a version comparison, so the two cannot drift apart again.
**Tests:** `TestDockerfileGoVersionSatisfiesGoMod` (MUST fail first).

### Task 4.2: Complete the Secrets Documentation
**Files:** Modify `docs/mac-mini-setup.md`
**Write-scope:** `docs/mac-mini-setup.md`
**Consumes:** `.env.example`, the `os.Getenv` call sites
**Produces:** A table of every variable the code reads, with the consequence of omitting each. Notes that values come from the existing Azure App Service configuration.
**Tests:** Cross-checked against `grep -rn "Getenv" --include="*.go"`.

### Task 4.3: Install Docker on the Mac Mini
**Produces:** A working Docker daemon set to start at login, with the login requirement understood — Docker Desktop needs a logged-in GUI session, so an unattended reboot leaves the daemon down while the runner is up.
**Tests:** `docker compose version` succeeds.
**Needs the user's hands:** admin install.

### Task 4.4: Dry Run the Stack by Hand
**Consumes:** Task 4.3, `~/tbd-bot-secrets/.env`
**Produces:** A real `docker compose up -d --build` on this machine — the same command `deploy.yml` runs, not a variant of it — with `tbd-bot` reaching `healthy`, Prometheus's scrape target `up`, and Grafana panels showing data.
**Tests:** `docker inspect --format='{{json .State.Health.Status}}' tbd-bot` returns `"healthy"`; `curl localhost:8080/metrics` returns metrics; `curl localhost:9090/api/v1/targets` reports the `tbd-bot` target as `up`.

### Task 4.5: Register the Runner and Lock It Down
**Consumes:** Task 4.3
**Produces:** A registered `self-hosted` runner installed as a service, fork-PR approval required in repository settings, and sleep disabled via `pmset`.
**Tests:** Runner shows Idle in **Settings -> Actions -> Runners**, and a `workflow_dispatch` job running `docker compose version` succeeds — proving the `launchd` service's minimal `PATH` can find the Docker CLI. The dry run in 4.4 does not prove this, because it uses the interactive shell's `PATH`.
**Needs the user's hands:** registration token from the GitHub UI, `sudo`.

### Task 4.6: Cut Over From Azure
**Consumes:** Tasks 4.4, 4.5
**Produces:** Azure Web App `tbdbot-cicd` stopped, bot confirmed offline in Discord, branch merged to `master`, deployment green on the self-hosted runner, bot back online. Azure left stopped rather than deleted so rollback is one click.
**Tests:** One slash command exercised end to end, with its RED metric visible in Grafana.
**Needs the user's hands:** Azure Portal, merge approval.

