# Mac Mini Self-Hosted Runner Setup Guide

Before merging the migration branch (`feature/mac-mini-migration`) into `master`, you **must** perform these manual setup steps on the Mac Mini. If you fail to do this, the GitHub Action will hang indefinitely waiting for a runner that does not exist, and the deployment will fail due to missing secrets.

## 1. Register the Mac Mini Runner
You must connect your Mac Mini to GitHub so it can listen for deployment webhooks.

1. Go to your GitHub repository in your web browser.
2. Navigate to **Settings -> Actions -> Runners**.
3. Click the green **New self-hosted runner** button.
4. Select **macOS** as the operating system and **ARM64** as the architecture (for M4).
5. Open the Terminal on your Mac Mini and run the commands exactly as GitHub provides them in the UI. They will look something like this:
   ```bash
   mkdir actions-runner && cd actions-runner
   curl -o actions-runner-osx-arm64-2.x.x.tar.gz -L https://github.com/actions/runner/releases/download/...
   tar xzf ./actions-runner-osx-arm64-2.x.x.tar.gz
   ./config.sh --url https://github.com/christiansoetanto/tbd-bot --token <YOUR_TOKEN>
   ```

## 2. Install the Runner as a Background Service
By default, if you just run `./run.sh`, the runner will die the moment you close the Terminal or restart the Mac Mini. We must install it as a persistent daemon.

While still inside the `actions-runner` directory on your Mac Mini, run:
```bash
sudo ./svc.sh install
sudo ./svc.sh start
```
The runner is now permanently listening in the background.

**Confirm the runner can see Docker.** The service runs under `launchd` with a minimal
`PATH`, not the `PATH` from your interactive shell, and Docker Desktop installs the CLI
into `/usr/local/bin` or `~/.docker/bin`. A dry run from Terminal proves nothing about
this. Verify it from the runner's own environment — the simplest check is a throwaway
`workflow_dispatch` job that runs `docker compose version`. If it cannot find `docker`,
add the directory to `PATH` in the runner's `.env` file (`actions-runner/.env`) and
restart the service with `sudo ./svc.sh stop && sudo ./svc.sh start`.

## 3. Create the Secrets File
To protect your Discord token from malicious pull requests, the `.env` file is intentionally excluded from the Git repository and the GitHub Actions workspace. The CI pipeline is hardcoded to copy the `.env` file from a secure, external directory on your Mac Mini (`~/tbd-bot-secrets/.env`).

1. Create the secure folder in your Mac Mini's home directory:
   ```bash
   mkdir -p ~/tbd-bot-secrets
   ```
2. Create and open the `.env` file:
   ```bash
   nano ~/tbd-bot-secrets/.env
   ```
3. Paste in your environment variables. Use `.env.example` in the repository as the
   authoritative list — every variable there must be present and filled in:

   | Variable | Purpose | Notes |
   |---|---|---|
   | `BOTTOKEN` | Discord bot token | Required; bot cannot connect without it. |
   | `TBDENV` | Firestore collection suffix and guild config selector | **Set to `prod`.** If unset, the bot reads guild config as prod but reads/writes Firestore collections named `users_`, `questions_`, `polls_`, `logs_` — empty collections, silently losing all existing data. |
   | `FIREBASE_CONFIG` | Service account JSON for Firestore | Single line. Required; no database access without it. |
   | `ROMCAL_API_FUNCTIONS_URL` | Liturgical calendar Cloud Function | Required; the calendar cron fails without it. |
   | `DEVMODE` | Dev toggle | `FALSE` in production. |
   | `PORT` | Health check / metrics port | `8080`; must match the compose port mapping. |
   | `GITHUBAPITOKEN` | GitHub API token used by the Office of Readings cron | Required for that cron. |
   | `GOMEMLIMIT` | Go soft memory limit | `200MiB`, kept under the 256M container limit. |
   | `GF_SECURITY_ADMIN_PASSWORD` | Grafana admin password | Choose a real password; the compose default is `admin`. |

   Copy the current values from the Azure App Service configuration for
   `tbdbot-cicd` so the new deployment points at the same Discord bot and the same
   Firestore collections.
4. Save and exit (`CTRL+O`, `Enter`, `CTRL+X`).

## 4. Lock Down Security (CRITICAL)
Running a self-hosted runner on your personal hardware gives GitHub Actions root-equivalent execution context on your Mac. If a stranger opens a malicious Pull Request, it could theoretically execute arbitrary code on your home network.

1. Go to your repository **Settings -> Actions -> General**.
2. Scroll down to **Fork pull request workflows from outside collaborators**.
3. Select **Require approval for all outside collaborators** (or disable it entirely).
4. Click Save.

## 5. Prevent macOS Sleep (Optional but Recommended)
The M4 Mac Mini is aggressive about power saving. If the Mac goes to sleep, Docker suspends, and the bot disconnects from Discord.

To prevent the Mac Mini from ever sleeping when plugged into the wall, run:
```bash
sudo pmset -c sleep 0 displaysleep 0 disksleep 0
```

## 6. Install Docker
The runner executes `docker compose` directly on this machine, so Docker must be
installed and running before the first deployment.

**Colima, not Docker Desktop.** Docker Desktop only runs while a user is logged in, so
a reboot that stops at the login screen leaves the daemon down while the runner is up.
Colima runs headless as a background service and comes back on its own after a reboot
or power loss.

1. Install Colima and the Docker CLI:
   ```bash
   brew install colima docker docker-compose
   ```
   The `docker-compose` formula supplies the Compose V2 CLI plugin that backs
   `docker compose`; it is not the old V1 standalone binary.

2. Start Colima and register it to start at boot:
   ```bash
   colima start --cpu 2 --memory 4 --disk 60
   brew services start colima
   ```

3. Confirm both the daemon and the compose plugin:
   ```bash
   docker version          # must show a Server section, not just Client
   docker compose version  # must succeed — deploy.yml uses this exact form
   ```
   If `docker compose` is missing but `docker-compose` works, change `deploy.yml` and
   `TestDeployWorkflow` in `main_test.go` together — the test asserts the literal
   command string.

4. Verify Colima survives a reboot before going further. Reboot the Mac Mini, do not
   log in, then SSH in and run `docker version`. If the server responds, the stack
   will recover from power loss unattended.

## 7. Cut Over From Azure
Production currently runs on the Azure Web App `tbdbot-cicd`, deployed by
`.github/workflows/master_tbdbot-cicd.yml`. Merging this branch deletes that workflow,
which stops future Azure deployments but **does not stop the container already
running there**.

There is no staging bot, so the dry run below uses production credentials — which
means the dry run *is* a live bot. Two instances sharing one `BOTTOKEN` produce
duplicate slash-command replies, duplicate cron posts (Office of Readings, Friday,
calendar), and concurrent writes to the same Firestore collections. **Azure must be
stopped before the dry run starts, not after it.**

`deploy.yml` only fires on push to `master`, so the dry run is what proves the image
builds, the secrets load, and the stack comes up. Do not skip it and let CI be the
first execution.

Run these in order. Pick a low-traffic window — the bot is offline between steps 1
and 3.

### 1. Stop Azure
Azure Portal -> App Service `tbdbot-cicd` -> **Stop**. Confirm in Discord that the bot
has gone offline before continuing. Leave the app stopped but **not deleted** — that
keeps rollback to a single **Start** click.

### 2. Dry run the stack by hand
From a checkout of `feature/mac-mini-migration` on the Mac Mini:

```bash
cp ~/tbd-bot-secrets/.env .env
docker compose up -d --build

# Must print "healthy" — retry for up to a minute while it starts.
docker inspect --format='{{json .State.Health.Status}}' tbd-bot

# Must print Prometheus metrics, not an error.
curl -s localhost:8080/metrics | head

# Must print "up". Prometheus reaches the bot as tbd-bot:8080 over the compose
# network; if this says "down", every Grafana panel will read "No data".
curl -s 'localhost:9090/api/v1/targets' | grep -o '"health":"[^"]*"'

# Grafana: log in and confirm the dashboard loaded and its panels have data.
open http://127.0.0.1:3000
```

Confirm the bot is back online in Discord and answers one slash command. Since this is
production, also confirm it is reading real data — a command that reads Firestore
proves `TBDENV` and `FIREBASE_CONFIG` are correct. Empty results mean `TBDENV` is
wrong and the bot is pointed at empty collections.

If anything fails, tear down with `docker compose down`, restart the Azure app to
restore service, and fix the problem before retrying.

### 3. Merge
Leave the dry-run stack running and merge `feature/mac-mini-migration` into `master`.
The runner picks up the push, injects the secrets, rebuilds, and recreates the
containers in place. Watch the Actions log through the health-check step.

### 4. Confirm and watch
Check that the bot is online, exercise one slash command, and confirm its RED metric
appears in Grafana. Over the next few days, watch that the cron jobs fire (Office of
Readings, Friday, calendar) and that memory stays under the 256M container limit.
