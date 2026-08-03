#!/usr/bin/env bash
#
# Pings an external dead-man's switch with the host's free disk space.
#
# This runs on the Mac itself rather than in the bot, because the container
# sees the Colima VM's disk and not the host's. On 2026-07-29 the host volume
# filled, Colima's LaunchAgent exited 1 while writing its config, and
# KeepAlive.SuccessfulExit=true meant launchd never retried it. Docker survived
# only because a manually started VM was already running. Nothing was watching
# the volume that filled.
set -euo pipefail

SECRETS="${HOME}/tbd-bot-secrets/.env"
VOLUME="${TBD_DISK_VOLUME:-/System/Volumes/Data}"
MIN_FREE_GB="${TBD_DISK_MIN_FREE_GB:-10}"

# The URL is a credential — holding it is enough to ping success and suppress
# the alert — so it stays in the secrets file. That file also holds a Firebase
# service account JSON blob, so the value is grepped out rather than sourced,
# which would execute the whole file.
if [ -z "${HEALTHCHECKS_DISK_PING_URL:-}" ] && [ -f "$SECRETS" ]; then
    RAW=$(grep -E '^HEALTHCHECKS_DISK_PING_URL=' "$SECRETS" | tail -1 | cut -d= -f2- || true)
    HEALTHCHECKS_DISK_PING_URL=$(printf '%s' "$RAW" | sed -e 's/^"//' -e "s/^'//" -e 's/"$//' -e "s/'$//")
fi

if [ -z "${HEALTHCHECKS_DISK_PING_URL:-}" ]; then
    echo "HEALTHCHECKS_DISK_PING_URL is not set; no disk check will be reported" >&2
    exit 0
fi

if [ -n "${TBD_DISK_FREE_GB:-}" ]; then
    FREE_GB="$TBD_DISK_FREE_GB"
else
    FREE_GB=$(df -g "$VOLUME" | awk 'NR==2 {print $4}')
fi

URL="${HEALTHCHECKS_DISK_PING_URL%/}"
BODY="${FREE_GB}GB free on ${VOLUME}, threshold ${MIN_FREE_GB}GB"

if [ "$FREE_GB" -lt "$MIN_FREE_GB" ]; then
    TARGET="${URL}/fail"
else
    TARGET="$URL"
fi

# A failed ping must not take the LaunchAgent down with it: KeepAlive only
# restarts after a clean exit, so a transient network error would stop the
# check permanently — the same trap that stopped Colima on 07-29.
if ! curl -fsS -m 10 --data "$BODY" "$TARGET" >/dev/null; then
    echo "failed to ping the disk check switch: $BODY" >&2
    exit 0
fi
