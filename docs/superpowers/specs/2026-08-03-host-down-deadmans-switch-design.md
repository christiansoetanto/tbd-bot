# Host-down dead-man's switch

Date: 2026-08-03
Status: approved, implementing

## Problem

Grafana, Prometheus and the bot all run on the same Mac Mini. Every alert built on
2026-08-03 is therefore blind to the one failure that takes the whole machine: power
loss, network loss, an unattended reboot, or Colima not coming back. When the host
goes dark the alerter goes dark with it, and silence reads as health — the same shape
as the 2026-08-01 outage, one layer up.

Nothing outside the house can poll the machine, so the signal has to be pushed out.

## Design

Two watchers, each covering the other's blind spot.

```
healthchecks.io  --watches-->  Mac Mini: power, network, OS, Colima, bot process
Grafana (local)  --watches-->  the pinger itself, gateway, scrape target
```

The bot pings healthchecks.io on a schedule. Silence past the grace period is the
alert. Both watchers deliver to the Discord webhook already provisioned in
`grafana/provisioning/alerting/contact-points.yml`, so there is one place to look.

### Bot heartbeat

`util/heartbeat.go`, driven by a `@every 1m` entry in the existing
`dbot.loadAllCronJobs`. `robfig/cron/v3` is already a dependency; no new one.

| Gateway state | Action |
|---|---|
| never connected (`LastHeartbeatAck` zero) | ping nothing |
| healthy | `POST https://hc-ping.com/<uuid>` |
| stale | `POST https://hc-ping.com/<uuid>/fail`, body = reason |

The suppressed first case is load-bearing. `GatewayHealthy()` is false until Ready
fires, and `/metrics` takes about 26s to answer after container start, so a tick that
lands during startup would `/fail` on every deploy. Alerts that cry wolf on every
deploy are alerts nobody reads. This mirrors the Docker `HEALTHCHECK --start-period`
and, unlike a timer, it needs no tuning: the gateway either has acknowledged a
heartbeat or it has not.

`/fail` carries the reason string from `GatewayHealthy()` so the Discord message says
`last Discord heartbeat acknowledged 4m12s ago` rather than `check failed`.

Check configuration: period 1m, grace 5m — roughly 6 minutes from host death to
Discord, comparable to the 5.5 minutes measured in the 08-03 fire drill.

### Watching the watcher

An unset ping URL must not look like health. This is the `GRAFANA_DISCORD_WEBHOOK_URL`
trap: provisioning succeeds, everything reads green, nothing is delivered. So the bot
exports its own pinger state and Grafana alerts on it.

```
tbd_bot_external_heartbeat_total{result="success|fail|error"}
tbd_bot_external_heartbeat_last_ping_timestamp_seconds
tbd_bot_external_heartbeat_enabled
```

The middle gauge tracks any *delivered* ping, healthy or failing, rather than only
healthy ones. Tracking success alone would make a gateway outage read as a broken
pinger as well, and `group_by: [alertname]` would then send that one outage to Discord
twice. `result="error"` — a transport failure or a non-2xx answer — is the case where
nothing was delivered and the external switch is genuinely blind.

An unset URL logs an error at startup and holds `enabled` at 0. Two rules cover the two
ways the switch can be silently absent, kept separate so each says one thing:
`tbd-bot-heartbeat-disabled` on `enabled < 1`, and `tbd-bot-heartbeat-not-delivering` on
a last ping older than 10 minutes. The second multiplies by
`(last_ping > bool 0)` so it stays at 0 until the first ping lands, which keeps a fresh
deploy from alerting during the window before the gateway connects.

Both use `noDataState: OK`, deliberately unlike the gateway and target rules.
`target-down` already owns "the bot is absent"; with `group_by: [alertname]` an
`Alerting` no-data state here would turn one outage into three Discord messages. These
rules exist only for "the bot is up but its pinger is not working".

### Colima and the disk: documented, not repaired

The 07-29 disk exhaustion killed the Colima LaunchAgent:

```
2026-07-29T12:45:14 level=fatal msg="error preparing config file: error writing yaml
file: open /Users/chris/.colima/default/colima.yaml: no space left on device"
```

Reading `~/Library/LaunchAgents/homebrew.mxcl.colima.plist` shows there is nothing to
repair, and that the error is not what it looked like. `ProgramArguments` is
`colima start -f` — foreground, so on a clean boot the process stays alive as intended.
The 11-second `already running, ignoring` loop happened only because Colima had been
started manually first: the foreground command found an existing VM, exited 0, and
`KeepAlive.SuccessfulExit=true` respawned it. That loop is the plist behaving as
configured, not a fault.

The real weakness is the same key. `SuccessfulExit=true` restarts the agent **only**
after a clean exit, so the 07-29 failure exit stopped it permanently. A transient
error at boot — a full disk, a Homebrew upgrade mid-flight — leaves Docker down with
no retry, which is precisely the condition this dead-man's switch exists to report.

Two things follow, and neither is a code change:

1. The current `error 1` state is session-local. It belongs to a login session alive
   since 07-29; on the next login launchd bootstraps the plist fresh. Restarting the
   service now would stop all three containers on the live deployment to fix nothing.
2. FileVault means a reboot needs someone at the keyboard regardless. That was decided
   on 07-29 and stands: the machine holds `BOTTOKEN` and the Firebase service account.

So the reboot path stays manual and stays documented in `docs/mac-mini-setup.md`. It is
recorded as untested, because it cannot be tested without a reboot.

Free space is not monitored either — see the clarification below. The heartbeat reports
a full disk after it has already taken the bot down, rather than warning before it does.

## Security

`HEALTHCHECKS_PING_URL` is a credential. Anyone holding it can ping success and
suppress the alert. It lives in
`~/tbd-bot-secrets/.env`, reaching a deploy only through the `cp` in `deploy.yml`, and
never appears in a tracked file. `docker-compose.yml` needs no change: the bot already
mounts the whole file via `env_file`.

## Testing

Written before the code, per the repo's standing practice.

- ping suppressed when the gateway has never connected
- success URL when healthy, `/fail` when stale
- reason string travels in the `/fail` body
- unset URL: no request, `enabled` gauge 0, error logged
- transport failure increments `result="error"` and does not panic
- URL construction does not double a trailing slash
- metrics registered exactly once
- a ping lands even when the caller's context has already expired
- `.env.example` documents the variable
- alerting provisioning contains the heartbeat rules with `noDataState: OK`
- no literal `hc-ping.com/<uuid>` in any tracked file

## Out of scope

- node_exporter and host metrics in Prometheus. Rejected as a larger component than
  this goal needs.
- Host disk monitoring. See the clarification below.
- Automatic reboot recovery. Blocked by FileVault, decided 07-29.

## Clarifications

**2026-08-03** — Host disk monitoring is dropped at the user's request. The design
originally paired the bot heartbeat with `scripts/host-disk-check.sh`, a user
LaunchAgent pinging a second check below a free-space threshold. It was built, tested
and then removed before merge.

The consequence is stated rather than hidden: nothing warns before the volume fills.
The 07-29 sequence can repeat — disk fills, Colima's LaunchAgent exits 1,
`KeepAlive.SuccessfulExit=true` declines to retry, Docker stays down. The difference
from 07-29 is that the outage no longer passes unnoticed for days: the heartbeat stops
and healthchecks.io alerts within about six minutes. Detection replaces prevention.
`df -h /System/Volumes/Data` was 22GB free at the time of writing, against the 10GB
threshold the dropped check would have used.
