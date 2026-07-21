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
3. Paste in your environment variables. It **must** contain these at minimum:
   ```env
   BOTTOKEN=your_actual_discord_bot_token
   GOMEMLIMIT=200MiB
   GF_SECURITY_ADMIN_PASSWORD=your_secure_grafana_password
   ```
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

---

### You are now ready to deploy!
You may now approve and merge the `feature/mac-mini-migration` branch into `master`. The Mac Mini will automatically detect the push, inject the secrets, build the Docker images, and launch the bot and monitoring stack.
