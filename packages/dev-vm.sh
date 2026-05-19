#!/usr/bin/env bash
set -euo pipefail

STATE_DIR="${DEV_VM_STATE_DIR:-/var/lib/dev-vms}"
REPO_DIR="${DEV_VM_FLAKE:-/home/joonas/work/joonas-sys-nixos}"

STATE_DIR="$(realpath -m "$STATE_DIR")"
LOCK_FILE="$STATE_DIR/state.lock"
LOCK_TIMEOUT="${DEV_VM_LOCK_TIMEOUT:-5}"
SSH_CONNECT_TIMEOUT=2
SSH_RETRY_INTERVAL=2
SSH_STARTUP_TIMEOUT=30

SLOTS=(
  dev-vm-1
  dev-vm-2
  dev-vm-3
  dev-vm-4
  dev-vm-5
)

ADDITIONAL_BIND_MOUNTS=(
  "/apps/opencode/:/home/joonas/.local/share/opencode/"
)

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

note() {
  printf '%s\n' "$*" >&2
}

usage() {
  cat >&2 <<EOF
usage:
  $0
  $0 up [PROJECT_DIR]
  $0 down [SLOT_OR_PROJECT]
  $0 down --all
  $0 status
  $0 ssh [SLOT_OR_PROJECT] [REMOTE_COMMAND...]
  $0 ssh -- [REMOTE_COMMAND...]
  $0 clean SLOT
  $0 clean --all

environment:
  DEV_VM_STATE_DIR   state directory, default /var/lib/dev-vms
  DEV_VM_FLAKE       flake directory, default /home/joonas/work/joonas-sys-nixos
EOF
}

usage_error() {
  printf 'error: %s\n\n' "$*" >&2
  usage
  exit 2
}

need_root() {
  if (( EUID != 0 )); then
    die "this script must be run as root"
  fi
}

refuse_root_ssh() {
  if (( EUID == 0 )); then
    die "dev-vm ssh must be run as the project user, not root"
  fi
}

ensure_state_dirs() {
  mkdir -p "$STATE_DIR/slots"
  chmod 0755 "$STATE_DIR" "$STATE_DIR/slots"
}

ensure_slot_dir() {
  local dir

  ensure_state_dirs
  dir="$(slot_dir "$1")"
  mkdir -p "$dir"
  chmod 0755 "$dir"
}

write_state_file() {
  local file="$1"
  local value="$2"

  printf '%s\n' "$value" >"$file"
  chmod 0644 "$file"
}

repair_slot_metadata_permissions() {
  local slot="$1"
  local file

  ensure_slot_dir "$slot"
  for file in \
    "$(project_file "$slot")" \
    "$(runner_file "$slot")" \
    "$(qemu_pid_file "$slot")" \
    "$(virtiofsd_pid_file "$slot")" \
    "$(started_file "$slot")"; do
    [[ -e "$file" ]] && chmod 0644 "$file"
  done
}

lock_state() {
  ensure_state_dirs
  exec 9>"$LOCK_FILE"
  flock -w "$LOCK_TIMEOUT" 9 || die "timed out waiting for state lock: $LOCK_FILE"
}

slot_ip() {
  case "$1" in
    dev-vm-1) printf '192.168.200.11' ;;
    dev-vm-2) printf '192.168.200.12' ;;
    dev-vm-3) printf '192.168.200.13' ;;
    dev-vm-4) printf '192.168.200.14' ;;
    dev-vm-5) printf '192.168.200.15' ;;
    *) return 1 ;;
  esac
}

slot_tap() {
  case "$1" in
    dev-vm-1) printf 'microvm-dev1' ;;
    dev-vm-2) printf 'microvm-dev2' ;;
    dev-vm-3) printf 'microvm-dev3' ;;
    dev-vm-4) printf 'microvm-dev4' ;;
    dev-vm-5) printf 'microvm-dev5' ;;
    *) return 1 ;;
  esac
}

slot_vsock() {
  case "$1" in
    dev-vm-1) printf '11' ;;
    dev-vm-2) printf '12' ;;
    dev-vm-3) printf '13' ;;
    dev-vm-4) printf '14' ;;
    dev-vm-5) printf '15' ;;
    *) return 1 ;;
  esac
}

is_slot() {
  local candidate="$1"
  local slot
  for slot in "${SLOTS[@]}"; do
    [[ "$candidate" == "$slot" ]] && return 0
  done
  return 1
}

slot_dir() {
  printf '%s/slots/%s' "$STATE_DIR" "$1"
}

root_dir() {
  printf '%s/root' "$(slot_dir "$1")"
}

runtime_dir() {
  printf '%s/run' "$(slot_dir "$1")"
}

slot_socket() {
  printf '/run/microvm-%s-root.sock' "$1"
}

project_file() {
  printf '%s/project' "$(slot_dir "$1")"
}

runner_file() {
  printf '%s/runner' "$(slot_dir "$1")"
}

qemu_pid_file() {
  printf '%s/qemu.pid' "$(slot_dir "$1")"
}

virtiofsd_pid_file() {
  printf '%s/virtiofsd.pid' "$(slot_dir "$1")"
}

started_file() {
  printf '%s/started-at' "$(slot_dir "$1")"
}

read_file() {
  local file="$1"
  [[ -f "$file" ]] && cat "$file"
}

assigned_project() {
  read_file "$(project_file "$1")"
}

pid_alive() {
  local pid="${1:-}"
  [[ "$pid" =~ ^[0-9]+$ ]] && kill -0 "$pid" 2>/dev/null
}

pid_matches_slot() {
  local pid="$1"
  local slot="$2"
  local cmdline

  pid_alive "$pid" || return 1
  cmdline="$(tr '\0' ' ' <"/proc/$pid/cmdline" 2>/dev/null || true)"
  [[ "$cmdline" == microvm@"$slot"* || "$cmdline" == *" -name $slot "* ]]
}

qemu_pid_for_slot() {
  local slot="$1"
  local pid

  pid="$(read_file "$(qemu_pid_file "$slot")" || true)"
  if pid_matches_slot "$pid" "$slot"; then
    printf '%s\n' "$pid"
    return 0
  fi

  pgrep -f "^microvm@${slot}( |$)" 2>/dev/null | head -n 1 || true
}

virtiofsd_pid_for_slot() {
  local slot="$1"
  local pid

  pid="$(read_file "$(virtiofsd_pid_file "$slot")" || true)"
  if pid_alive "$pid"; then
    printf '%s\n' "$pid"
    return 0
  fi

  pgrep -f "$(slot_socket "$slot")" 2>/dev/null | head -n 1 || true
}

slot_running() {
  local pid
  pid="$(qemu_pid_for_slot "$1")"
  [[ -n "$pid" ]]
}

lexical_absolute_path() {
  local path="$1"
  local old_ifs part
  local -a parts normalized

  if [[ "$path" != /* ]]; then
    path="${PWD:-$(pwd -L)}/$path"
  fi

  old_ifs="$IFS"
  IFS='/'
  read -r -a parts <<<"$path"
  IFS="$old_ifs"

  normalized=()
  for part in "${parts[@]}"; do
    case "$part" in
      ""|".")
        ;;
      "..")
        if ((${#normalized[@]} > 0)); then
          unset 'normalized[-1]'
        fi
        ;;
      *)
        normalized+=("$part")
        ;;
    esac
  done

  if ((${#normalized[@]} == 0)); then
    printf '/\n'
  else
    old_ifs="$IFS"
    IFS='/'
    printf '/%s\n' "${normalized[*]}"
    IFS="$old_ifs"
  fi
}

caller_user() {
  if [[ -n "${SUDO_USER:-}" && "${SUDO_USER:-}" != "root" ]]; then
    printf '%s\n' "$SUDO_USER"
  else
    id -un
  fi
}

caller_uid() {
  if [[ "${SUDO_UID:-}" =~ ^[0-9]+$ ]]; then
    printf '%s\n' "$SUDO_UID"
  else
    id -u "$(caller_user)"
  fi
}

caller_gid() {
  if [[ "${SUDO_GID:-}" =~ ^[0-9]+$ ]]; then
    printf '%s\n' "$SUDO_GID"
  else
    id -g "$(caller_user)"
  fi
}

caller_home() {
  local user

  user="$(caller_user)"
  awk -F: -v user="$user" '$1 == user { print $6; exit }' /etc/passwd
}

prefer_user_visible_work_path() {
  local path="$1"
  local home visible_root visible_root_real path_real suffix visible

  home="$(caller_home || true)"
  [[ -n "$home" ]] || {
    printf '%s\n' "$path"
    return
  }

  visible_root="$home/work"
  [[ -d "$visible_root" ]] || {
    printf '%s\n' "$path"
    return
  }

  case "$path" in
    "$visible_root"|"$visible_root"/*)
      printf '%s\n' "$path"
      return
      ;;
  esac

  visible_root_real="$(realpath -m "$visible_root" 2>/dev/null || true)"
  path_real="$(realpath -m "$path" 2>/dev/null || true)"
  [[ -n "$visible_root_real" && -n "$path_real" ]] || {
    printf '%s\n' "$path"
    return
  }

  case "$path_real" in
    "$visible_root_real")
      visible="$visible_root"
      ;;
    "$visible_root_real"/*)
      suffix="${path_real#"$visible_root_real"/}"
      visible="$visible_root/$suffix"
      ;;
    *)
      printf '%s\n' "$path"
      return
      ;;
  esac

  if [[ -d "$visible" ]]; then
    printf '%s\n' "$visible"
  else
    printf '%s\n' "$path"
  fi
}

project_path() {
  prefer_user_visible_work_path "$(lexical_absolute_path "$1")"
}

validate_project() {
  local project="$1"

  [[ -d "$project" ]] || die "project directory does not exist: $project"
  project="$(project_path "$project")"
  [[ -d "$project" ]] || die "project directory does not exist: $project"

  [[ "$project" != "/" ]] || die "refusing to mount / as a project"
  [[ "$project" != "/nix/store" && "$project" != /nix/store/* ]] || die "refusing to use /nix/store as a project"
  [[ "$project" != *$'\n'* && "$project" != *$'\t'* ]] || die "project path must not contain newlines or tabs"

  printf '%s\n' "$project"
}

validate_slot_root() {
  local root="$1"
  case "$root" in
    "$STATE_DIR"/slots/dev-vm-*/root) ;;
    *) die "internal safety check failed for slot root: $root" ;;
  esac
}

mounts_under_root() {
  local root="$1"
  findmnt -rn --raw -o TARGET 2>/dev/null \
    | awk -v prefix="${root}/" 'index($0, prefix) == 1 { print length($0) "\t" $0 }' \
    | sort -rn \
    | cut -f2-
}

cleanup_mounts() {
  local slot="$1"
  local root
  local target
  root="$(root_dir "$slot")"

  [[ -d "$root" ]] || return 0
  validate_slot_root "$root"

  while IFS= read -r target; do
    [[ -n "$target" ]] || continue
    umount "$target" 2>/dev/null || umount -l "$target" 2>/dev/null || true
  done < <(mounts_under_root "$root")
}

reset_root() {
  local slot="$1"
  local root
  root="$(root_dir "$slot")"

  validate_slot_root "$root"
  mkdir -p "$root"
  find "$root" -mindepth 1 -maxdepth 1 -exec rm -rf -- {} +
}

prepare_store_overlay() {
  local slot="$1"
  local dir root upper work mountpoint

  dir="$(slot_dir "$slot")"
  root="$(root_dir "$slot")"
  upper="$dir/store-upper"
  work="$dir/store-work"
  mountpoint="$root/nix/store"

  mkdir -p "$mountpoint" "$upper" "$work"

  if mountpoint -q "$mountpoint"; then
    return 0
  fi

  mount -t overlay "devvm-${slot}-store" \
    -o "lowerdir=/nix/store,upperdir=${upper},workdir=${work}" \
    "$mountpoint"
}

bind_project() {
  local slot="$1"
  local project="$2"
  local dest

  dest="$(root_dir "$slot")$project"
  mkdir -p "$dest"
  mount --bind "$project" "$dest"
  mount -o remount,rw,bind "$dest"
}

mkdir_p_for_caller() {
  local path="$1"
  local uid gid current i
  local -a missing

  uid="$(caller_uid)"
  gid="$(caller_gid)"

  current="$path"
  while [[ "$current" != "/" && ! -e "$current" ]]; do
    missing+=("$current")
    current="$(dirname "$current")"
  done

  [[ -d "$current" ]] || die "mount point parent is not a directory: $current"

  for ((i = ${#missing[@]} - 1; i >= 0; i--)); do
    mkdir "${missing[$i]}"
    chown "$uid:$gid" "${missing[$i]}"
  done

  [[ -d "$path" ]] || die "mount point is not a directory: $path"
}

bind_additional_mounts() {
  local slot="$1"
  local mount_spec source target dest

  for mount_spec in "${ADDITIONAL_BIND_MOUNTS[@]}"; do
    source="${mount_spec%%:*}"
    target="${mount_spec#*:}"

    [[ -d "$source" ]] || die "additional bind mount source does not exist: $source"
    [[ "$source" == /* ]] || die "additional bind mount source must be absolute: $source"
    [[ "$target" == /* ]] || die "additional bind mount target must be absolute: $target"
    [[ "$target" != "/" ]] || die "refusing to bind mount over guest root"

    dest="$(root_dir "$slot")$target"
    mkdir_p_for_caller "$dest"
    mount --bind "$source" "$dest"
    mount -o remount,rw,bind "$dest"
  done
}

prepare_slot_root() {
  local slot="$1"
  local project="$2"

  cleanup_mounts "$slot"
  reset_root "$slot"
  prepare_store_overlay "$slot"
  bind_project "$slot" "$project"
  bind_additional_mounts "$slot"
}

build_runner() {
  local slot="$1"
  nix build --no-link --print-out-paths "${REPO_DIR}#nixosConfigurations.${slot}.config.microvm.declaredRunner"
}

wait_for_socket() {
  local socket="$1"
  local pid="$2"
  local label="$3"
  local i

  for i in $(seq 1 100); do
    [[ -S "$socket" ]] && return 0
    if ! pid_alive "$pid"; then
      return 1
    fi
    sleep 0.1
  done

  die "timed out waiting for $label socket: $socket"
}

wait_for_exit() {
  local pid="$1"
  local i

  for i in $(seq 1 100); do
    pid_alive "$pid" || return 0
    sleep 0.1
  done

  return 1
}

stop_pid() {
  local pid="$1"
  local label="$2"

  pid_alive "$pid" || return 0
  kill "$pid" 2>/dev/null || true
  if ! wait_for_exit "$pid"; then
    note "forcing $label pid $pid to stop"
    kill -9 "$pid" 2>/dev/null || true
    wait_for_exit "$pid" || true
  fi
}

stop_slot() {
  local slot="$1"
  local clear_assignment="${2:-yes}"
  local dir runner qemu_pid virtiofsd_pid tap qmp

  dir="$(slot_dir "$slot")"
  runner="$(read_file "$(runner_file "$slot")" || true)"
  qemu_pid="$(qemu_pid_for_slot "$slot")"
  virtiofsd_pid="$(virtiofsd_pid_for_slot "$slot")"
  tap="$(slot_tap "$slot")"
  qmp="$(runtime_dir "$slot")/${slot}.sock"

  if [[ -n "$qemu_pid" ]]; then
    if [[ -n "$runner" && -x "$runner/bin/microvm-shutdown" ]]; then
      ( cd "$(runtime_dir "$slot")" && timeout 15 "$runner/bin/microvm-shutdown" >/dev/null 2>&1 ) || true
      wait_for_exit "$qemu_pid" || true
    fi
    stop_pid "$qemu_pid" "qemu"
  fi

  if [[ -n "$virtiofsd_pid" ]]; then
    stop_pid "$virtiofsd_pid" "virtiofsd"
  fi

  ip link delete "$tap" >/dev/null 2>&1 || true
  rm -f "$(slot_socket "$slot")" "$qmp"

  cleanup_mounts "$slot"
  reset_root "$slot"

  rm -f "$(qemu_pid_file "$slot")" "$(virtiofsd_pid_file "$slot")"

  if [[ "$clear_assignment" == "yes" ]]; then
    rm -f "$(project_file "$slot")" "$(runner_file "$slot")" "$(started_file "$slot")"
  fi

  ensure_slot_dir "$slot"
}

recover_stale() {
  local slot project

  for slot in "${SLOTS[@]}"; do
    project="$(assigned_project "$slot" || true)"
    if [[ -n "$project" ]] && ! slot_running "$slot"; then
      note "recovering stale state for $slot ($project)"
      stop_slot "$slot"
    fi
  done
}

slot_for_project() {
  local project="$1"
  local slot assigned

  for slot in "${SLOTS[@]}"; do
    assigned="$(assigned_project "$slot" || true)"
    [[ "$assigned" == "$project" ]] || continue
    printf '%s\n' "$slot"
    return 0
  done
}

free_slot() {
  local slot assigned

  for slot in "${SLOTS[@]}"; do
    assigned="$(assigned_project "$slot" || true)"
    if [[ -z "$assigned" ]] && ! slot_running "$slot"; then
      printf '%s\n' "$slot"
      return 0
    fi
  done
}

write_assignment() {
  local slot="$1"
  local project="$2"
  local runner="$3"
  local qemu_pid="$4"
  local virtiofsd_pid="$5"

  ensure_slot_dir "$slot"
  write_state_file "$(project_file "$slot")" "$project"
  write_state_file "$(runner_file "$slot")" "$runner"
  write_state_file "$(qemu_pid_file "$slot")" "$qemu_pid"
  write_state_file "$(virtiofsd_pid_file "$slot")" "$virtiofsd_pid"
  write_state_file "$(started_file "$slot")" "$(date --iso-8601=seconds)"
}

ensure_prerequisites() {
  command -v nix >/dev/null || die "nix is not in PATH"
  command -v mount >/dev/null || die "mount is not in PATH"
  command -v findmnt >/dev/null || die "findmnt is not in PATH"
  command -v flock >/dev/null || die "flock is not in PATH"
  command -v setsid >/dev/null || die "setsid is not in PATH"
  command -v ip >/dev/null || die "ip is not in PATH"
  [[ -d /nix/store ]] || die "/nix/store is missing"
}

start_slot() {
  local slot="$1"
  local project="$2"
  local dir run root_socket runner virtiofsd_pid qemu_pid qmp
  local start_ok="no"

  dir="$(slot_dir "$slot")"
  run="$(runtime_dir "$slot")"
  root_socket="$(slot_socket "$slot")"

  trap '[[ "$start_ok" == "yes" ]] || stop_slot "$slot"' RETURN

  ensure_slot_dir "$slot"
  mkdir -p "$run"

  note "building runner for $slot"
  runner="$(build_runner "$slot")"
  ln -sfn "$runner" "$dir/current"

  note "preparing root export for $slot"
  prepare_slot_root "$slot" "$project"

  rm -f "$root_socket" "$run/${slot}.sock"

  note "starting virtiofsd for $slot"
  (
    exec 9>&-
    cd "$run"
    exec setsid "$runner/bin/virtiofsd-run"
  ) >"$dir/virtiofsd.log" 2>&1 &
  virtiofsd_pid="$!"

  if ! wait_for_socket "$root_socket" "$virtiofsd_pid" "virtiofsd"; then
    stop_slot "$slot"
    die "virtiofsd for $slot exited before creating $root_socket; see $dir/virtiofsd.log"
  fi

  note "starting qemu for $slot"
  (
    exec 9>&-
    cd "$run"
    exec setsid "$runner/bin/microvm-run"
  ) >"$dir/console.log" 2>&1 &
  qemu_pid="$!"

  qmp="$run/${slot}.sock"
  if ! wait_for_socket "$qmp" "$qemu_pid" "qmp"; then
    stop_slot "$slot"
    die "qemu for $slot exited before creating $qmp; see $dir/console.log"
  fi

  write_assignment "$slot" "$project" "$runner" "$qemu_pid" "$virtiofsd_pid"

  printf '%s\n' "started $slot"
  printf '  project: %s\n' "$project"
  printf '  ip:      %s\n' "$(slot_ip "$slot")"
  printf '  vsock:   %s\n' "$(slot_vsock "$slot")"
  printf '  logs:    %s/{console.log,virtiofsd.log}\n' "$dir"

  start_ok="yes"
  trap - RETURN
}

cmd_up() {
  local project="$1"
  local slot assigned

  project="$(validate_project "$project")"
  ensure_prerequisites
  recover_stale

  slot="$(slot_for_project "$project" || true)"
  if [[ -n "$slot" ]]; then
    if slot_running "$slot"; then
      printf '%s is already running in %s\n' "$project" "$slot"
      return 0
    fi
    stop_slot "$slot"
  else
    slot="$(free_slot || true)"
  fi

  [[ -n "$slot" ]] || die "no free dev VM slots"

  assigned="$(assigned_project "$slot" || true)"
  [[ -z "$assigned" ]] || die "$slot is still assigned to $assigned"

  start_slot "$slot" "$project"
}

resolve_slot() {
  local target="$1"
  local project assigned slot

  if is_slot "$target"; then
    printf '%s\n' "$target"
    return 0
  fi

  if [[ -d "$target" ]]; then
    project="$(project_path "$target")"
  else
    project="$target"
  fi

  for slot in "${SLOTS[@]}"; do
    assigned="$(assigned_project "$slot" || true)"
    [[ "$assigned" == "$project" ]] || continue
    printf '%s\n' "$slot"
    return 0
  done

  die "unknown slot or project: $target"
}

cmd_down() {
  local target="$1"
  local slot

  if [[ "$target" == "--all" ]]; then
    for slot in "${SLOTS[@]}"; do
      stop_slot "$slot"
    done
    return 0
  fi

  slot="$(resolve_slot "$target")"
  stop_slot "$slot"
}

cmd_clean() {
  local target="$1"
  local slot dir

  if [[ "$target" == "--all" ]]; then
    for slot in "${SLOTS[@]}"; do
      slot_running "$slot" && die "$slot is running; stop it before cleaning"
    done
    for slot in "${SLOTS[@]}"; do
      cmd_clean "$slot"
    done
    return 0
  fi

  slot="$(resolve_slot "$target")"
  slot_running "$slot" && die "$slot is running; stop it before cleaning"

  stop_slot "$slot"
  dir="$(slot_dir "$slot")"
  rm -rf -- "$dir/store-upper" "$dir/store-work"
  ensure_slot_dir "$slot"
}

cmd_status() {
  local slot project state started store_mount project_mount qemu_pid virtiofsd_pid

  printf '%-9s %-8s %-15s %-7s %s\n' \
    "slot" "state" "ip" "mounts" "project"
  printf '%-9s %-8s %-15s %-7s %s\n' \
    "---------" "--------" "---------------" "-------" "-------"

  for slot in "${SLOTS[@]}"; do
    repair_slot_metadata_permissions "$slot"

    project="$(assigned_project "$slot" || true)"
    started="$(read_file "$(started_file "$slot")" || true)"
    qemu_pid="$(qemu_pid_for_slot "$slot")"
    virtiofsd_pid="$(virtiofsd_pid_for_slot "$slot")"

    if [[ -n "$qemu_pid" ]]; then
      state="running"
    elif [[ -n "$project" || -n "$virtiofsd_pid" ]]; then
      state="stale"
    else
      state="free"
    fi

    store_mount="store:no"
    mountpoint -q "$(root_dir "$slot")/nix/store" 2>/dev/null && store_mount="store:yes"

    project_mount="project:no"
    if [[ -n "$project" ]]; then
      mountpoint -q "$(root_dir "$slot")$project" 2>/dev/null && project_mount="project:yes"
    fi

    if [[ -n "$started" ]]; then
      project="${project} (${started})"
    fi

    printf '%-9s %-8s %-15s %-7s %s\n' \
      "$slot" "$state" "$(slot_ip "$slot")" \
      "${store_mount#store:}/${project_mount#project:}" "${project:-}"
  done

  cat <<EOF

legend:
  tap = microvm-dev<slot number>
  vsock = IP last octet
EOF
}

run_ssh_with_startup_retry() {
  local slot="$1"
  shift

  local deadline status sleep_for
  deadline=$((SECONDS + SSH_STARTUP_TIMEOUT))

  while true; do
    ssh \
      -o ConnectTimeout="$SSH_CONNECT_TIMEOUT" \
      -o ConnectionAttempts=1 \
      "$@" && return 0
    status="$?"

    if (( status != 255 )); then
      return "$status"
    fi

    if (( SECONDS >= deadline )); then
      die "timed out after ${SSH_STARTUP_TIMEOUT}s waiting for SSH to $slot"
    fi

    sleep_for="$SSH_RETRY_INTERVAL"
    if (( SECONDS + sleep_for > deadline )); then
      sleep_for=$((deadline - SECONDS))
    fi

    note "SSH to $slot is not ready yet; retrying in ${sleep_for}s"
    sleep "$sleep_for"
  done
}

cmd_ssh() {
  local target="$1"
  shift

  local slot cid project quoted_project
  slot="$(resolve_slot "$target")"
  cid="$(slot_vsock "$slot")"

  if (($# == 0)); then
    project="$(assigned_project "$slot" || true)"
    [[ -n "$project" ]] || die "$slot has no assigned project"
    printf -v quoted_project '%q' "$project"
    run_ssh_with_startup_retry "$slot" \
      -t "vsock%${cid}" "cd ${quoted_project} && exec \"\$SHELL\" -l"
    exit "$?"
  fi

  run_ssh_with_startup_retry "$slot" "vsock%${cid}" "$@"
  exit "$?"
}

cmd_default() {
  local project slot

  refuse_root_ssh

  project="$(validate_project ".")"
  slot="$(slot_for_project "$project" || true)"

  if [[ -z "$slot" ]] || ! slot_running "$slot"; then
    command -v sudo >/dev/null || die "sudo is required to start a dev VM"
    sudo dev-vm up "$project"
  fi

  cmd_ssh "$project"
}

main() {
  local command="${1:-}"
  shift || true

  case "$command" in
    "")
      cmd_default
      ;;
    up)
      [[ $# -le 1 ]] || usage_error "unrecognized arguments: ${*:2}"
      need_root
      lock_state
      cmd_up "${1:-.}"
      ;;
    down)
      [[ $# -le 1 ]] || usage_error "unrecognized arguments: ${*:2}"
      need_root
      lock_state
      cmd_down "${1:-.}"
      ;;
    status)
      [[ $# -eq 0 ]] || usage_error "unrecognized arguments: $*"
      need_root
      lock_state
      cmd_status
      ;;
    ssh)
      refuse_root_ssh
      if [[ "${1:-}" == "--" ]]; then
        shift
        cmd_ssh "." "$@"
      elif [[ $# -eq 0 ]]; then
        cmd_ssh "."
      else
        cmd_ssh "$@"
      fi
      ;;
    clean)
      [[ $# -gt 0 ]] || usage_error "missing SLOT or --all"
      [[ $# -eq 1 ]] || usage_error "unrecognized arguments: ${*:2}"
      need_root
      lock_state
      cmd_clean "$1"
      ;;
    -h|--help)
      usage
      ;;
    *)
      usage_error "unrecognized arguments: $command${*:+ $*}"
      exit 2
      ;;
  esac
}

main "$@"
