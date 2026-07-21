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
