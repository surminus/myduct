#!/usr/bin/env bash
#
# Shut the machine down once it has been idle for a while.
#
# Idle means nobody is logged in over ssh or mosh, no Claude Code session is
# running, and the one minute load average is below the threshold. A detached
# zellij or tmux session running a build keeps the load up, so it counts as
# busy even with nobody attached. Claude needs its own check because a session
# waiting on the API or on me to approve something sits at near zero load.
#
# Run from a systemd timer every five minutes. Each idle run increments a
# counter; a busy run resets it. Once the counter reaches the limit we power
# off. With a five minute interval and a limit of six that is thirty minutes.

set -euo pipefail

STATE_FILE=/run/claude-box-autostop.count
IDLE_RUNS_BEFORE_STOP=${IDLE_RUNS_BEFORE_STOP:-6}
LOAD_THRESHOLD=${LOAD_THRESHOLD:-0.5}

logged_in() {
  if [ -n "$(who)" ]; then
    return 0
  fi

  if pgrep -x mosh-server >/dev/null; then
    return 0
  fi

  # ssh sessions without a tty, like scp or an editor's remote backend, do
  # not show up in who, but they do leave a per-session sshd process owned
  # by the user. Recent OpenSSH calls it sshd-session, so no exact match.
  if pgrep -u laura sshd >/dev/null 2>&1; then
    return 0
  fi

  return 1
}

claude_running() {
  pgrep -u laura -x claude >/dev/null 2>&1
}

busy_load() {
  # load is a gawk builtin, hence the awkward variable name
  local current
  current=$(cut -d' ' -f1 /proc/loadavg)
  awk -v current="$current" -v threshold="$LOAD_THRESHOLD" 'BEGIN { exit !(current >= threshold) }'
}

count=$(cat "$STATE_FILE" 2>/dev/null || true)
case "$count" in
  ''|*[!0-9]*) count=0 ;;
esac

if logged_in || claude_running || busy_load; then
  echo 0 > "$STATE_FILE"
  exit 0
fi

count=$((count + 1))
echo "$count" > "$STATE_FILE"

if [ "$count" -ge "$IDLE_RUNS_BEFORE_STOP" ]; then
  logger -t claude-box-autostop "idle for $count checks, shutting down"
  systemctl poweroff
fi
