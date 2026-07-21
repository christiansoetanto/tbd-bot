# Spec: Heroku to Mac Mini Migration
**Lifecycle status:** planned

## Goal
Migrate the `tbd-bot` deployment architecture from Heroku to an M4 Mac Mini home server. This involves optimizing the Docker container, establishing deep application metrics (Prometheus/Grafana), and setting up automated CI/CD via a local GitHub Runner.

## Non-goals
- Migrating to a Cloud VPS or Serverless architecture.
- Modifying the core Go application logic beyond adding Prometheus metrics and graceful shutdown logic.
- Establishing remote/external access to the Grafana dashboard.

## Constraints
- The Mac Mini must maintain persistent internet access for the Discord WebSocket session.
- Secrets (`BOTTOKEN`, etc.) must be stored securely in a local `.env` file on the Mac Mini, and securely injected into the GitHub Actions workspace.
- Deployments must be automated via GitHub Actions. **Self-hosted runner security must be enforced** by strictly tying workflows to the `main` branch and disabling workflows from fork Pull Requests via GitHub Repository Settings.
- The system and application timezone must remain in UTC.
- The bot must avoid Discord Gateway bans by implementing a graceful `SIGTERM` shutdown that calls `discordgo.Session.Close()`.

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

## Acceptance criteria
- **Wave 1 (Metrics & Lifecycle):** The Go codebase is updated to expose `/metrics` on port 8080. A graceful `SIGTERM` trap closes the Discord session. RED metrics are manually injected into event handlers.
- **Wave 2 (Docker & Monitoring Setup):** A multi-stage `Dockerfile` is implemented with a native `HEALTHCHECK`. A `docker-compose.yml` is created to launch `tbd-bot` (with `GOMEMLIMIT` and memory limits), `prometheus` (with 14d/1GB limits), and `grafana` (bound to `127.0.0.1`). Baseline JSON dashboards are provided.
- **Wave 3 (Pipeline Setup):** A `.github/workflows/deploy.yml` action is configured with `cancel-in-progress: true`, secure `.env` workspace injection, and a true Docker health-check polling step.

**Roles touched:** none
