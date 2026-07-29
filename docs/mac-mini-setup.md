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

**Install the service as the same user that installed Colima** (`chris`). `svc.sh`
installs the runner as a LaunchDaemon; two pieces of the Docker setup are user-scoped
and break if the daemon runs as anyone else:

- `~/.docker/config.json`, which is what makes `docker compose` resolve at all.
- The Colima VM and its socket, under `~/.colima/`.

**Confirm the runner can see Docker.** The daemon gets a minimal `PATH` — not the one
from your interactive shell — and Homebrew's CLI lives in `/opt/homebrew/bin`. A dry
run from Terminal proves nothing about this. Verify from the runner's own environment
with a throwaway `workflow_dispatch` job:

```yaml
- run: |
    whoami
    docker version
    docker compose version
```

`whoami` must print the user that owns Colima. If `docker` is not found, add
`PATH=/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin` to `actions-runner/.env` and
restart with `sudo ./svc.sh stop && sudo ./svc.sh start`. If `docker` is found but
`docker compose` is not, the daemon is not reading the right `~/.docker/config.json` —
check `whoami` before touching `PATH` again.

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

**Colima, not Docker Desktop.** Colima is headless — no GUI app, no menu-bar process,
lower idle overhead — which is what a server wants.

**It does not, however, survive an unattended reboot on this machine.** `brew services
start colima` installs a *LaunchAgent* at `~/Library/LaunchAgents/homebrew.mxcl.colima.plist`,
and LaunchAgents only run inside a user login session. FileVault is enabled on this Mac
Mini, so the disk needs a password at boot and automatic login is not possible. After a
reboot or power loss:

- The GitHub runner **is** up — `sudo ./svc.sh install` creates a LaunchDaemon, which
  runs at boot with no login.
- Colima and the containers are **down** until someone logs in.
- Deployments in that window fail with `cannot connect to the Docker daemon`.

The options are to leave it as is and log in after a reboot, or to disable FileVault and
enable automatic login, which trades disk encryption for unattended recovery. This
machine holds `BOTTOKEN` and the Firebase service account JSON, so **keeping FileVault
on is the better trade** — power loss is rare, and the recovery is one login.

Once someone logs in, recovery is automatic the rest of the way: Colima starts via the
LaunchAgent and the containers come back on their own, because every service in
`docker-compose.yml` is `restart: unless-stopped`.

**After any reboot, the order is: log in, confirm `docker version` answers, then re-run
any deploy that failed while the machine was down.** A deploy that fires before login
hits a dead daemon, and the failure reads like a code problem when it is not.

1. Install Colima and the Docker CLI:
   ```bash
   brew install colima docker docker-compose
   ```
   The `docker-compose` formula supplies the Compose V2 CLI plugin that backs
   `docker compose`; it is not the old V1 standalone binary.

2. **Point the Docker CLI at the Homebrew plugin directory.** Homebrew installs the
   Compose plugin somewhere the CLI does not search by default, so `docker compose`
   fails with `docker: unknown command: docker compose` until this file exists:
   ```bash
   mkdir -p ~/.docker && cat > ~/.docker/config.json <<'JSON'
   {
     "cliPluginsExtraDirs": [
       "/opt/homebrew/lib/docker/cli-plugins"
     ]
   }
   JSON
   docker compose version   # must now print a version
   ```
   This file lives in the invoking user's home directory. **The runner must run as the
   same user**, or it will not see this config and `deploy.yml` will fail at the deploy
   step even though the command works in your terminal.

3. Start Colima and register it to start at boot:
   ```bash
   colima start --cpu 2 --memory 4 --disk 60
   brew services start colima
   ```

4. Confirm both the daemon and the compose plugin:
   ```bash
   docker version          # must show a Server section, not just Client
   docker compose version  # must succeed — deploy.yml uses this exact form
   ```
   If `docker compose` is missing but `docker-compose` works, change `deploy.yml` and
   `TestDeployWorkflow` in `main_test.go` together — the test asserts the literal
   command string.

5. Know the reboot behaviour before going further. Reboot the Mac Mini and, without
   logging in, try `docker version` over SSH — it will fail, for the reasons above.
   Log in, wait a moment, and run it again; the server should respond and the
   containers should be back. That second result is the recovery path to rely on.

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

**Measure how long `/metrics` takes to answer.** This is the first time `dbot.Init`
runs in a container, and the HTTP server starts only after it finishes registering
slash commands with Discord. The healthcheck allows 60s (`--start-period`) and
`deploy.yml` gives up at 120s; both numbers were chosen without a measurement:

```bash
time until curl -sf localhost:8080/metrics >/dev/null; do sleep 1; done
```

A few seconds means the current bounds are fine. Anything approaching 60s means raise
both `--start-period` and the workflow timeout before relying on this pipeline.

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
