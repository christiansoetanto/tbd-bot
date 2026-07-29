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

1. Install Docker Desktop for Apple Silicon (or Colima + the Docker CLI).
2. Set Docker Desktop to **start at login**, otherwise a reboot leaves the bot down
   until someone opens the app.
3. Confirm which compose command exists on this machine — the workflow must match it:
   ```bash
   docker compose version   # Compose V2, the plugin form used by deploy.yml
   docker-compose --version # Compose V1, the standalone binary
   ```

## 7. Dry Run Before Merging (CRITICAL)
`deploy.yml` only runs on push to `master`, so without a dry run the first real
execution of this stack happens against production Discord. Run it by hand first,
from a checkout of `feature/mac-mini-migration`:

```bash
cp ~/tbd-bot-secrets/.env .env
docker compose up -d --build
docker inspect --format='{{json .State.Health.Status}}' tbd-bot
curl -s localhost:8080/metrics | head
open http://127.0.0.1:3000   # Grafana; dashboard should already be provisioned
```

**Use a staging token and `TBDENV=staging` for the dry run if one is available.** With
production values, this dry run *is* a second live bot — see the cutover step below
before starting it with production credentials.

Tear the dry run down with `docker compose down` once verified.

## 8. Cut Over From Azure
Production currently runs on the Azure Web App `tbdbot-cicd`, deployed by
`.github/workflows/master_tbdbot-cicd.yml`. Merging this branch deletes that workflow,
which stops future Azure deployments but **does not stop the container already
running there**. Two instances sharing one `BOTTOKEN` produce duplicate slash-command
replies, duplicate cron posts (Office of Readings, Friday, calendar), and concurrent
writes to the same Firestore collections.

Order of operations:

1. Finish the dry run above and tear it down.
2. Stop the Azure Web App `tbdbot-cicd` (Azure Portal -> the App Service -> **Stop**).
3. Confirm in Discord that the bot has gone offline.
4. Merge `feature/mac-mini-migration` into `master`. The runner picks up the push,
   injects the secrets, builds the images, and starts the stack.
5. Confirm the bot is back online, then exercise one slash command and check Grafana
   for the corresponding RED metric.
6. Leave the Azure app stopped (not deleted) until the Mac Mini has run cleanly for a
   few days, so rollback is a single **Start** click.

## 9. Deploy
Once steps 1-8 are done, merge the branch. The Mac Mini detects the push, injects the
secrets, builds the Docker images, and launches the bot and monitoring stack.
