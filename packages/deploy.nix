{ pkgs }:

pkgs.writeShellApplication {
  name = "deploy.sh";

  runtimeInputs = with pkgs; [
    coreutils
    findutils
    fzf
    libnotify
    sway
  ];

  text = ''
    set -euo pipefail

    if [ ! -t 0 ]; then
      swaymsg "exec \$term -e \"$0\""
      exit 0
    fi

    deployer_dir="/sysroot/apps/deployer"
    deployer_bin="$deployer_dir/rel/deployer_linux-amd64"
    env_file="$deployer_dir/env-prod.env"

    cd "$deployer_dir"

    service="''${1:-}"
    if [ -z "$service" ]; then
      service="$(
        find ./deployments/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' \
          | sort \
          | fzf --prompt='Service: '
      )"
    fi

    # shellcheck source=/dev/null
    source "$env_file"

    "$deployer_bin" deploy "$service" ""

    notify-send "Deployed: $service"
  '';
}
