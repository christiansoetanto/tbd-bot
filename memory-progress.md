# Memory Progress — tbd-bot

- 2026-07-22: Completed Galdr setup and repo architecture analysis. Created `docs/agents/galdr.md`, `CLAUDE.md`, `docs/agents/roles-and-journeys.md`, `memory.md`, and `memory-progress.md`.
- 2026-07-22: Completed Task 1.1: Prometheus HTTP Handler. Added `/metrics` endpoint to HTTP server and verified with `main_test.go`.
- 2026-07-22: Completed Task 1.2: Business Metrics Counters. Created `util/metrics.go` (`tbd_bot_qa_moves_total`, `tbd_bot_users_vetted_total`), injected counter increments in `sd_command.go` and `reaction_add.go`, initialized in `dbot/init.go`, and verified with `util/metrics_test.go` and `main_test.go`.
- 2026-07-22: Completed Task 2.1: Multi-stage Dockerfile. Updated `Dockerfile` with `golang:1.22-alpine` builder stage, `CGO_ENABLED=0`, static binary build, and native Docker `HEALTHCHECK` using `wget` against `http://localhost:8080/metrics`. Added unit test `TestDockerfile_MultiStageAndHealthcheck` in `main_test.go` and verification script `scripts/verify_dockerfile.sh`.
- 2026-07-22: Completed Task 2.3: Documentation. Updated `README.md` with Grafana local access (`127.0.0.1:3000`) and GitHub repository settings instructions for securing self-hosted runners against fork PRs. Verified with `TestDocumentation_GrafanaAndRunnerSecurity` in `main_test.go`.
- 2026-07-22: Completed Task 2.4: Baseline Dashboard JSON. Created `grafana/provisioning/dashboards/bot-dashboard.json` visualizing Prometheus RED metrics (Rate, Errors, Duration) and expanded business counters. Verified valid JSON parsing with `python3 -m json.tool` and `TestGrafanaDashboardJSON` in `main_test.go`.
- 2026-07-22: Completed Task 3.1: Self-Hosted Runner Deploy Action. Modified `.github/workflows/deploy.yml` with `self-hosted` runner target, `push` on `main`, `concurrency: cancel-in-progress: true`, secure `.env` secret injection from `~/tbd-bot-secrets/.env`, `docker-compose up -d --build`, and post-deployment health check polling using `docker inspect --format='{{json .State.Health.Status}}'`. Validated with `actionlint` and `TestDeployWorkflow` in `main_test.go`.

next: Wave 3 VPS Migration Plan completed.



- DECISION [Infrastructure] Hosting Platform → Biznet Gio Ubuntu VPS (treating as normal SSH, no proprietary cloud tools)
- DECISION [Infrastructure] Architecture Model → Persistent WebSocket listener via VPS (Serverless rejected due to Discord MessageCreate dependencies)
- DECISION [Deployment] Containerization → Multi-stage Dockerfile (Alpine base) to reduce image size and improve security
- DECISION [Monitoring] Metrics Exposure → Use the existing port 8080 HTTP server to expose the /metrics endpoint.
- DECISION [Monitoring] Metrics Scope → Track both operational and business-specific metrics (e.g., Q&A moves, vetting counts, detailed cron statuses).
- DECISION [Deployment] CI/CD Pipeline → Use GitHub Self-Hosted Runner on the Mac Mini to automate deployments locally.
- DECISION [Monitoring] RED Metrics Interception → Implement a custom Go middleware wrapper for Discordgo handlers to automatically track Latency (Duration), Success Rate (Errors), and QPS (Rate).
- DECISION [Monitoring] Data Persistence → Use local Docker volumes (e.g., `./data/prometheus`) to persist metrics and dashboards across reboots.
- DECISION [Monitoring] Dashboards as Code → Use Grafana provisioning to load dashboard layouts automatically from JSON files stored in the repository.
- DECISION [Infrastructure] Remote Access → Keep Grafana local-only for security; document local access steps in README.
- DECISION [Security] GitHub Runner → Disable fork PR workflow execution in GitHub Repo Settings to prevent self-hosted runner hijack.
- DECISION [Infrastructure] Docker Volumes → Use Docker named volumes instead of bind mounts for Prometheus/Grafana to avoid macOS permission loops.
- DECISION [Deployment] CI Validation → Enforce actionlint for YAML tests and add a post-deploy container health check to prevent silent failures.
- DECISION [Reliability] Staff SRE Audits → Enforce 1GB/14d Prometheus retention, container memory limits, CGO_ENABLED=0, native Docker HEALTHCHECK, CI concurrency limits, and secure `.env` injection.
- DECISION [Logic] Timezone → Keep cron and system timezone as UTC (rejected SRE tzdata suggestion).
- DECISION [Reliability] Chaos Monkey Audits → Enforce graceful `Session.Close()`, `GOMEMLIMIT` injection, `127.0.0.1` Grafana binding, and `wget` healthchecks.
- DECISION [Logic] RED Metrics Interception → Abandon generic reflection middleware; manually inject timer decorators into specific handler registrations.
