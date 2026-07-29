# Memory Progress — tbd-bot

- 2026-07-22: Completed Galdr setup and repo architecture analysis. Created `docs/agents/galdr.md`, `CLAUDE.md`, `docs/agents/roles-and-journeys.md`, `memory.md`, and `memory-progress.md`.
- 2026-07-22: Wrote Epic: Mac Mini Deployment & Monitoring (waves 1-3). Containerized the bot, implemented Prometheus RED metrics and business counters, orchestrated via `docker-compose`, built an IaC Grafana dashboard, and wrote a deploy workflow targeting a GitHub self-hosted runner. Full specs at `docs/specs/2026-07-22-vps-migration.md`. Compacted architecture details into `memory.md`. **Code complete on `feature/mac-mini-migration`, not merged and not deployed.**
- 2026-07-29: Added wave 4 (host setup and cutover) after verifying actual state on the Mac Mini itself (`hw.model` = `Mac16,10`). Three findings: the builder image `golang:1.22-alpine` was older than `go.mod`'s `go 1.25.0` so the Docker build could never have succeeded (fixed, now asserted by `TestDockerfileGoVersionSatisfiesGoMod`); `docs/mac-mini-setup.md` documented only 3 of the 8 required env vars, omitting `TBDENV` and `FIREBASE_CONFIG` (fixed); and the plan had no cutover step, so merging would have left the Azure instance running alongside the new one on the same `BOTTOKEN`. Docker, the runner, and `~/tbd-bot-secrets/.env` do not exist on this machine yet.

next: Wave 4 — install Docker on the Mac Mini (4.3), then dry-run the stack by hand (4.4) before merging. See `docs/mac-mini-setup.md` sections 6-9.

