# Repository Memory — tbd-bot

## System Overview
`tbd-bot` is a Go-based Discord bot designed for multi-guild management (Servus Dei, Capital Mindset, Terra Sancta) with specialized Catholic liturgical features and strict community vetting/moderation tooling.

## Architecture & Layers
- **`main.go`**: Entry point. Parses environment variables (`DEVMODE`, `BOTTOKEN`, `PORT`, `TBDENV`, `FIREBASE_CONFIG`), initializes config, ZeroLog logging, Firestore, Discord session, and an HTTP health check server (port 8080 default).
- **`config/`**: Guild-specific configurations (Staging vs Production) mapped by `GuildId`. Controls feature toggles (`RegisteredFeature`), channel/role mappings, and thresholds (e.g., question limiters).
- **`domain/`**: Core types and constants (`FeatureKey`, `ReligionRoleKey`, `Question`, `Poll`, `Option`, `Voter`).
- **`dbot/init.go`**: Handles bot startup by bulk overwriting Discord Slash Commands based on guild feature toggles and registering `robfig/cron/v3` jobs. 
- **`dbot/handler/`**: Discord interaction handlers (slash commands, message components, keyword detection, reaction events).
- **`provider/dbms/firestore.go`**: Database layer. Appends the `TBDENV` suffix to collections (e.g., `users_staging`, `questions_prod`) for `users`, `logs`, `questions`, and `polls`.
- **`util/`**: Helpers for scraping, API fetching, Markdown conversion, and Discord Embed building.

## Deep Feature Mechanics
1. **Vetting & Verification (`/sdverify`)**: 
   - A strict state machine: Adds `ApprovedUser` + specific `ReligionRole` while actively stripping `VettingQuestioning` and all `Vetting` roles.
   - Detects invalid vetting responses via keywords and regex (e.g., checking for the secret INRI code).
2. **Liturgical Calendar (`util/calendar.go`)**: 
   - Fetches from a custom Romcal API via a Google Cloud Function (`ROMCAL_API_FUNCTIONS_URL`).
   - Identifies Holy Days of Obligation, Feasts, and Solemnities. (Note: The logic to `@mention` Latin Catholics for HDOs is currently hard-disabled in code).
3. **Office of Readings (`util/office_of_readings.go`)**: 
   - Scrapes `ibreviary.com` using `goquery`. Extracts the "SECOND READING" node, converts HTML to Markdown, strips out the Responsory, and chunks text to fit Discord's 3000-character embed limit.
4. **Q&A Management & Mover (`dbot/handler/reaction_add.go`)**: 
   - Triggered manually by a moderator reacting to a question. 
   - The bot searches up to 3000 messages in the discussion channel to find the answer, builds a jump URL, posts an embed to `#answered-questions`, and **deletes** the original question.
5. **CM Question Limiter (`dbot/handler/message.go`)**: 
   - Limits users in Capital Mindset from asking questions too frequently (tracked via Firestore timestamps).
   - Violators receive a DM and have their question deleted (bypassed by `UnlimitedRoleIds`).
6. **Polls & Voting**: Interactive Discord UI buttons with state persisted in Firestore.

## Environment Variables
- `BOTTOKEN`: Discord Bot Token (Required).
- `DEVMODE`: Boolean string (`true`/`false`).
- `TBDENV`: Environment identifier (`staging` or `prod`).
- `PORT`: HTTP health check port (Default `8080`).
- `FIREBASE_CONFIG`: Service account JSON string for Firebase/Firestore.
- `ROMCAL_API_FUNCTIONS_URL`: Cloud Function URL for calendar data.
- `GOMEMLIMIT`: Soft memory limit for Go garbage collection (required for Docker).
- `GF_SECURITY_ADMIN_PASSWORD`: Admin password for Grafana.

## Deployment & Monitoring (Mac Mini)
- **Deployment**: Automated via GitHub Actions (`.github/workflows/deploy.yml`) targeting a local `self-hosted` runner on an M4 Mac Mini.
- **Infrastructure**: Containerized via a multi-stage `Dockerfile` (`golang:1.22-alpine` with `CGO_ENABLED=0` and `wget` healthcheck) and orchestrated via `docker-compose.yml`.
- **Secrets Management**: The `.env` file is excluded from Git and injected directly from `~/tbd-bot-secrets/.env` on the Mac Mini runner to prevent unauthorized PR exfiltration.
- **Data Safety**: Uses Docker named volumes (`prometheus_data`, `grafana_data`) to survive reboots without macOS permission loops. Prometheus is strictly limited to 14 days or 1GB retention.
- **Monitoring (Prometheus/Grafana)**:
  - Custom HTTP `/metrics` server running on port 8080.
  - **RED Metrics**: Manual timer decorators wrap Discord event handlers (`tbd_bot_handler_duration_seconds`, `tbd_bot_handler_requests_total`).
  - **Business Counters**: Extensively tracks `tbd_bot_cron_executions_total`, `tbd_bot_qa_moves_total`, `tbd_bot_users_vetted_total`, etc.
  - **Dashboards as Code**: Grafana is provisioned via IaC to auto-load datasources and `bot-dashboard.json`.
  - **Lifecycle**: Uses a graceful `SIGINT`/`SIGTERM` OS trap to explicitly call `discordgo.Session.Close()` upon container shutdown to prevent Gateway disconnect bans.
