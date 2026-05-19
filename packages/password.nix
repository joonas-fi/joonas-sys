{ pkgs }:

let
  script = pkgs.writeShellApplication {
    name = "password.sh";

    runtimeInputs = with pkgs; [
      fuzzel
      wl-clipboard
    ];

    text = ''
      set -eu

      # check if not called from terminal => open in terminal window
      if [ ! -t 0 ]; then
        # the `-e` (instead of `--command`) is lowest common denominator for (alacritty, ghostty)
        swaymsg "exec \$term -e \"$0\""
        exit 0
      fi

      cd /apps/Desktop/Salasanoja

      file=$(find . | fzf)

      cat "$file"
      read -r # enter to exit reading the file. (`-r` to satisfy shellcheck)
    '';
  };
  desktopItem = pkgs.makeDesktopItem {
    name = "password";
    desktopName = "Find password";
    genericName = "Custom Script";
    exec = "${script}/bin/password.sh";
    icon = "application-x-executable";
    terminal = false;
    categories = [ "Utility" ];
  };
in
pkgs.symlinkJoin {
  name = "password";
  paths = [
    script
    desktopItem
  ];
}
