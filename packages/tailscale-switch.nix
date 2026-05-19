{ pkgs }:

let
  script = pkgs.writeShellApplication {
    name = "tailscale-switch.sh";

    runtimeInputs = with pkgs; [
      tailscale
      fzf
    ];

    text = ''
      set -eu

      # check if not called from terminal => open in terminal window
      if [ ! -t 0 ]; then
        # the `-e` (instead of `--command`) is lowest common denominator for (alacritty, ghostty)
        swaymsg "exec \$term -e \"$0\""
        exit 0
      fi

      nickname=$(echo -e "home\nssrc\n" | fzf);

      exec sudo tailscale switch "$nickname"
    '';
  };
  desktopItem = pkgs.makeDesktopItem {
    name = "tailscale-switch";
    desktopName = "Tailscale switch tailnet";
    genericName = "Custom Script";
    exec = "${script}/bin/tailscale-switch.sh";
    icon = "application-x-executable";
    terminal = false;
    categories = [ "Utility" ];
  };
in
pkgs.symlinkJoin {
  name = "tailscale-switch";
  paths = [
    script
    desktopItem
  ];
}
