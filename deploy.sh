#!/usr/bin/env bash
#
# deploy.sh — one-command deploy of the SSH TUI (ssh-tui) to whoop-vm.
#
# whoop-vm is the tailnet node behind georgenijo.com: the small Oracle Cloud
# box that runs the "ssh-tui" systemd --user service which `ssh
# georgenijo.com` actually lands on. This script cross-compiles the TUI for
# that box, ships the binary over, and restarts the service in place.
#
# Usage:
#   ./deploy.sh
#
# What it does:
#   1. Builds ./ssh-tui for linux/amd64 (static, no cgo), stamping the
#      binary with the current year-month via -ldflags -X main.buildStamp=.
#   2. Verifies the build compiles/runs locally (go build already did this;
#      this script does not run the app, just builds it).
#   3. scp's the binary to whoop-vm as ssh-tui.new (never overwrites the
#      live binary directly, so a failed transfer can't half-write it).
#   4. Over ssh: moves ssh-tui.new into place, chmod +x's it, restarts the
#      user-level systemd unit, waits for it to settle, and confirms it's
#      active.
#   5. Compares the local and remote md5sum of the binary to make sure the
#      byte-for-byte artifact that shipped is the one now running, and
#      fails loudly (non-zero exit, clear error) on any mismatch or on the
#      service failing to come back up.
#
# Requires: go, scp, ssh, md5sum, and SSH access as ubuntu@whoop-vm (an
# entry in ~/.ssh/config or working tailnet DNS/IP is assumed).

set -euo pipefail

REMOTE_USER="ubuntu"
REMOTE_HOST="100.114.248.101" # whoop-vm, over the tailnet
REMOTE_DIR="ssh-tui"          # relative to the remote user's $HOME
SERVICE="ssh-tui"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="${SCRIPT_DIR}/ssh-tui"

BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "${BUILD_DIR}"' EXIT
BIN="${BUILD_DIR}/ssh-tui"

# Lowercase "month year" (e.g. "july 2026") to match the site's now-page stamp.
STAMP="$(date +'%B %Y' | tr '[:upper:]' '[:lower:]')"

echo "==> [1/6] Building ssh-tui for linux/amd64 (buildStamp=${STAMP})"
(
  cd "${SRC_DIR}"
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -X 'main.buildStamp=${STAMP}'" \
    -o "${BIN}" \
    .
)
echo "    built ${BIN}"

echo "==> [2/6] Computing local md5sum"
LOCAL_MD5="$(md5sum "${BIN}" | awk '{print $1}')"
echo "    local md5: ${LOCAL_MD5}"

echo "==> [3/6] Copying binary to ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/ssh-tui.new"
scp "${BIN}" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/ssh-tui.new"

echo "==> [4/6] Installing binary and restarting ${SERVICE} on whoop-vm"
REMOTE_OUT="$(ssh "${REMOTE_USER}@${REMOTE_HOST}" bash -s -- "${REMOTE_DIR}" "${SERVICE}" <<'REMOTE_SCRIPT'
set -euo pipefail
remote_dir="$1"
service="$2"

if [ -f "${remote_dir}/ssh-tui" ]; then
  echo "    (remote) keeping rollback copy at ${remote_dir}/ssh-tui.prev"
  cp "${remote_dir}/ssh-tui" "${remote_dir}/ssh-tui.prev"
fi

echo "    (remote) mv ${remote_dir}/ssh-tui.new -> ${remote_dir}/ssh-tui"
mv "${remote_dir}/ssh-tui.new" "${remote_dir}/ssh-tui"

echo "    (remote) chmod +x ${remote_dir}/ssh-tui"
chmod +x "${remote_dir}/ssh-tui"

echo "    (remote) systemctl --user restart ${service}"
systemctl --user restart "${service}"

echo "    (remote) waiting 2s for the service to settle"
sleep 2

status="$(systemctl --user is-active "${service}" || true)"
md5="$(md5sum "${remote_dir}/ssh-tui" | awk '{print $1}')"

echo "STATUS=${status}"
echo "MD5=${md5}"
REMOTE_SCRIPT
)"

echo "${REMOTE_OUT//$'\n'/$'\n'    }" | sed '1s/^/    /'

REMOTE_STATUS="$(echo "${REMOTE_OUT}" | grep -E '^STATUS=' | tail -1 | cut -d= -f2)"
REMOTE_MD5="$(echo "${REMOTE_OUT}" | grep -E '^MD5=' | tail -1 | cut -d= -f2)"

echo "==> [5/6] Verifying remote service is active"
if [[ "${REMOTE_STATUS}" != "active" ]]; then
  echo "!!! DEPLOY FAILED: ${SERVICE} is not active on whoop-vm (systemctl --user is-active reported '${REMOTE_STATUS}')" >&2
  exit 1
fi
echo "    ${SERVICE} is active"

echo "==> [6/6] Verifying local and remote binaries match (md5sum)"
echo "    local:  ${LOCAL_MD5}"
echo "    remote: ${REMOTE_MD5}"
if [[ "${REMOTE_MD5}" != "${LOCAL_MD5}" ]]; then
  echo "!!! DEPLOY FAILED: md5sum mismatch between local build and remote binary" >&2
  echo "!!! local=${LOCAL_MD5} remote=${REMOTE_MD5}" >&2
  exit 1
fi

echo "==> Deploy complete: ${SERVICE}@whoop-vm running buildStamp=${STAMP} (md5 ${LOCAL_MD5})"
