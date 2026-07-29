# Spec: Azure App Service to Mac Mini Migration
**Lifecycle status:** code complete on `feature/mac-mini-migration`; host setup and cutover pending

## Goal
Migrate the `tbd-bot` deployment architecture from its current host to an M4 Mac Mini home server. This involves optimizing the Docker container, establishing deep application metrics (Prometheus/Grafana), and setting up automated CI/CD via a local GitHub Runner.

The current production host is the Azure Web App `tbdbot-cicd`, deployed by `.github/workflows/master_tbdbot-cicd.yml`. (Earlier drafts of this spec said Heroku; the leftover `Procfile` is a relic of that older host and is no longer in use.)

## Non-goals
- Migrating to a Cloud VPS or Serverless architecture.
- Modifying the core Go application logic beyond adding Prometheus metrics and graceful shutdown logic.
- Establishing remote/external access to the Grafana dashboard.

## Constraints
- The Mac Mini must maintain persistent internet access for the Discord WebSocket session.
- Secrets (`BOTTOKEN`, etc.) must be stored securely in a local `.env` file on the Mac Mini, and securely injected into the GitHub Actions workspace.
- Deployments must be automated via GitHub Actions. **Self-hosted runner security must be enforced** by strictly tying workflows to the `master` branch and disabling workflows from fork Pull Requests via GitHub Repository Settings.
- The system and application timezone must remain in UTC.
- The bot must avoid Discord Gateway bans by implementing a graceful `SIGTERM` shutdown that calls `discordgo.Session.Close()`.
- Only one instance may run against the production `BOTTOKEN` at a time. The Azure Web App must be stopped before the Mac Mini stack starts with production credentials; two live instances cause duplicate command replies, duplicate cron posts, and concurrent writes to the same Firestore collections.
- The `.env` on the Mac Mini must carry every variable the code reads, not just the deployment-specific ones. `TBDENV` in particular is the Firestore collection suffix — unset, the bot silently reads and writes empty `users_` / `questions_` / `polls_` / `logs_` collections while still loading production guild config.

## Decision log

| # | Decision | Choice |
|---|---|---|
| 1 | Hosting Platform | M4 Mac Mini Home Server. |
| 2 | Metrics Method | Attach `/metrics` endpoint to the existing port 8080 HTTP server. |
| 3 | Metrics Scope | Track standard operational metrics AND business-specific counters (Q&A moves, vetting, cron runs). |
| 4 | Containerization | Multi-stage Dockerfile (Alpine base) compiled with `CGO_ENABLED=0` and a `wget`-based HEALTHCHECK. |
| 5 | Deployment Pipeline | GitHub Self-Hosted Runner installed on the Mac Mini. Includes concurrency cancellation to prevent race conditions. |
| 6 | RED Metrics | Use manual Prometheus timer decorators injected into specific handlers to automatically measure Latency, Success Rate, and QPS (avoiding slow Go reflection). |
| 7 | Data Persistence | Use **Docker Named Volumes** for Prometheus and Grafana. Prometheus limited to 14d/1GB retention. Bot bounded by Docker memory limits and a matching `GOMEMLIMIT`. |
| 8 | Dashboards as Code | Use Grafana Provisioning to auto-load Dashboards and Datasources from JSON files checked into the Git repository. |
| 9 | Remote Access | Grafana remains local-only for maximum security by explicitly binding to `127.0.0.1:3000`. Access instructions will be documented in the README. |
| 10 | CI Validation | Enforce `actionlint` for YAML testing, implement a native Docker `HEALTHCHECK`, and poll `docker inspect` post-deploy. |
| 11 | Cutover | Manual, old-off-then-new-on. Stop the Azure Web App, confirm the bot is offline in Discord, then dry-run and merge. Azure stays stopped (not deleted) for a few days so rollback is one click. |
| 12 | First execution | The stack is run by hand on the Mac Mini before merging. `deploy.yml` fires only on push to `master`, so without a dry run its first execution would also be the first real test of the image build, compose command, and secrets. |
| 13 | Container runtime | **Colima**, not Docker Desktop. Colima runs headless as a background service and recovers from an unattended reboot or power loss; Docker Desktop needs a logged-in GUI session. |
| 14 | Dry-run credentials | Production. There is no staging bot token, so the dry run is itself the live cutover — Azure is stopped first, and the bot is offline for the duration of the build. |

## Acceptance criteria
- **Wave 1 (Metrics & Lifecycle):** The Go codebase is updated to expose `/metrics` on port 8080. A graceful `SIGTERM` trap closes the Discord session. RED metrics are manually injected into event handlers.
- **Wave 2 (Docker & Monitoring Setup):** A multi-stage `Dockerfile` is implemented with a native `HEALTHCHECK`. A `docker-compose.yml` is created to launch `tbd-bot` (with `GOMEMLIMIT` and memory limits), `prometheus` (with 14d/1GB limits), and `grafana` (bound to `127.0.0.1`). Baseline JSON dashboards are provided.
- **Wave 3 (Pipeline Setup):** A `.github/workflows/deploy.yml` action is configured with `cancel-in-progress: true`, secure `.env` workspace injection, and a true Docker health-check polling step.
- **Wave 4 (Host Setup & Cutover):** Docker, the self-hosted runner, and `~/tbd-bot-secrets/.env` exist on the Mac Mini. The compose command in `deploy.yml` matches what is actually installed. The stack has been built and run by hand on the Mac Mini and reported healthy. The Azure Web App is stopped, the branch is merged, and the bot is confirmed online with metrics visible in Grafana.

## Clarifications

### 2026-07-29
- The prior host is Azure App Service (`tbdbot-cicd`), not Heroku. The spec's original framing was wrong; the `Procfile` predates Azure and is unused.
- Cutover is explicitly in scope. It was missing from the original plan, which ended at the pipeline being configured and left two live instances as an unhandled outcome.
- The builder image must track `go.mod`. The pinned `golang:1.22-alpine` was older than `go 1.25.0` and the Docker build could not have succeeded; the version relationship is now asserted by `TestDockerfileGoVersionSatisfiesGoMod` rather than a literal tag.
- The Docker-related tests are text assertions over file contents, not executions. They are not evidence that the image builds or that the stack runs; only the manual dry run on the Mac Mini is.
- `deploy.yml` invokes Compose V2 (`docker compose`), not the standalone `docker-compose` binary, which Docker Desktop no longer ships. The dry run must use the identical command, or it would step around the most likely failure.
- The dry run cannot prove the runner works. The `launchd` service has a minimal `PATH` and will not necessarily find the Docker CLI, so that is verified separately from the runner's own environment.
- Runtime is Colima rather than Docker Desktop, chosen for unattended reboot recovery.
- The dry run uses production credentials because no staging bot exists. This collapses "dry run" and "cutover" into one sequence: Azure stops first, the bot is offline while the image builds, and the dry run doubles as the go-live. The acceptance check therefore includes reading real Firestore data, since a wrong `TBDENV` would otherwise look like a healthy but empty bot.

**Roles touched:** none
