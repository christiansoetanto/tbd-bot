# TBD-BOT
Because I couldn't think of a name at time of writing the code.

## Features
- Vetting system
  - Verification and role assigning (/sdverify)
  - Vetting response's secret code detector (the INRI code)
  - Welcome message
- Daily Catholic
  - Office of Readings
  - Liturgical celebration for today
  - Friday abstinence memes
- Q&A
  - #religious-questions, #religious-discussions, #religious-discussions-2, and #answered-questions system where, as the name implies, answered questions will be moved to #answered-questions so that unanswered questions can have more exposure.
## Monitoring & Grafana Access
Grafana is bound to `127.0.0.1:9400` on the host (container port 3000), so it is reachable from the Mac Mini itself and
nowhere else. Dashboards are auto-loaded from
`grafana/provisioning/dashboards/bot-dashboard.json`.

- **URL:** `http://127.0.0.1:9400` — use the IP, not `localhost`, which may resolve to IPv6 and hit a different app
- **Username:** `admin`
- **Password:** whatever `GF_SECURITY_ADMIN_PASSWORD` was set to in `.env`.
  Self-registration is disabled (`GF_USERS_ALLOW_SIGN_UP=false`).

### Viewing it from another machine
Don't change the port binding to `0.0.0.0` — that publishes a Grafana admin login to
your whole network. Tunnel over SSH instead:

```bash
ssh -L 9400:127.0.0.1:9400 <user>@<mac-mini>
```

Then open `http://127.0.0.1:9400` on the local machine.

### Changing the admin password
`GF_SECURITY_ADMIN_PASSWORD` is only read when Grafana initializes its database — the
first boot with an empty `grafana_data` volume. After that, editing the variable and
restarting has no effect, because Grafana reads the password from its own database.
To change it later:

```bash
docker exec -it grafana grafana-cli admin reset-admin-password 'new-password'
```

Update `.env` to match afterwards, or the file will disagree with reality.

## Alerting
Alerts are provisioned as code in `grafana/provisioning/alerting/` and delivered to a Discord
webhook. A webhook is used instead of email because it needs no credentials, is not subject to
spam filtering, and does not depend on `BOTTOKEN` or on the bot process — so it still delivers
when the bot itself is what is down.

Set `GRAFANA_DISCORD_WEBHOOK_URL` in **`~/tbd-bot-secrets/.env`** (Discord: Server Settings →
Integrations → Webhooks). Point it at a channel only you can read and set that channel to
"All Messages" for a phone push. The URL is a credential — anyone holding it can post to that
channel — so it is read from the environment and never committed.

Which file matters: `deploy.yml` runs `cp ~/tbd-bot-secrets/.env .env` before composing, so the
secrets file is canonical and the repo `.env` is a deploy artefact. Compose only ever reads the
repo `.env` — it has no knowledge of the secrets file — so if you run compose by hand without
that copy, the repo `.env` must carry the value too. Grafana is the unforgiving case: it has no
`env_file` and receives the URL purely through `${GRAFANA_DISCORD_WEBHOOK_URL}` interpolation,
which silently resolves to an empty string when unset. Alerting then provisions cleanly and
delivers nowhere.

Three rules fire into it:

| Rule | Fires when | No-data behaviour |
|------|-----------|-------------------|
| `tbd-bot-gateway-stale` | No Discord heartbeat acknowledged for over 5 minutes | Alerting |
| `tbd-bot-target-down` | Prometheus cannot scrape the bot | Alerting |
| `tbd-bot-external-api-failing` | Any GitHub or Discord API call failed in the last 15 minutes | OK |

The first two alert on missing data on purpose. Every signal that failed during the 2026-08-01
outage failed by going quiet rather than by going red, so a rule that treats absence as health
would reproduce exactly that.

### Health is the gateway, not the HTTP server
`/health` returns 503 once the Discord gateway stops acknowledging heartbeats, and the Docker
`HEALTHCHECK` probes it. It used to probe `/metrics`, which answers 200 off the HTTP server
alone — that is why the container reported `healthy` for 33 hours while the bot was invisible
in Discord. `/metrics` itself stays unconditional so Prometheus keeps scraping through an
outage; a metrics endpoint that fails during an incident destroys the series needed to alert
on it.

### Removing a provisioned alert rule
Deleting the file does **not** delete the rule. Grafana copies provisioned rules into its own
database and keeps them after the file is gone — the same trap as `GF_SECURITY_ADMIN_PASSWORD`
above. Removal takes an explicit directive in a provisioning file:

```yaml
apiVersion: 1
deleteRules:
  - orgId: 1
    uid: the-rule-uid
```

Restart Grafana, confirm the rule is gone, then delete the directive file.

## Self-Hosted Runner Security
To protect the local host environment and self-hosted runner from arbitrary code execution via external Fork Pull Requests:
1. Open your repository on GitHub.
2. Navigate to **Settings** -> **Actions** -> **General**.
3. Under **Fork pull request workflows from outside collaborators**, choose **Require approval for all outside collaborators** (or disable fork PR execution on self-hosted runners).
4. Ensure workflow jobs targeting `runs-on: self-hosted` are tied strictly to the `master` branch.

## [![Repography logo](https://images.repography.com/logo.svg)](https://repography.com) / Recent activity [![Time period](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_badge.svg)](https://repography.com)
[![Timeline graph](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_timeline.svg)](https://github.com/christiansoetanto/tbd-bot/commits)
[![Issue status graph](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_issues.svg)](https://github.com/christiansoetanto/tbd-bot/issues)
[![Pull request status graph](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_prs.svg)](https://github.com/christiansoetanto/tbd-bot/pulls)
[![Trending topics](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_words.svg)](https://github.com/christiansoetanto/tbd-bot/commits)
[![Top contributors](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_users.svg)](https://github.com/christiansoetanto/tbd-bot/graphs/contributors)
[![Activity map](https://images.repography.com/26965455/christiansoetanto/tbd-bot/recent-activity/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/ZHy8hl1p33C5CZb9geQ1DTotvDo6YiFPQ9owxkhS1qU_map.svg)](https://github.com/christiansoetanto/tbd-bot/commits)



## [![Repography logo](https://images.repography.com/logo.svg)](https://repography.com) / Structure
[![Structure](https://images.repography.com/26965455/christiansoetanto/tbd-bot/structure/XSKu8NuvUO-wMkvxmAk0-Sh04dEgjpwow1r37BcSWVk/FNR-8bU2tOwB9W1WX1pp7RKuhSjJagCinaWXKcfjiXk_table.svg)](https://github.com/christiansoetanto/tbd-bot)

