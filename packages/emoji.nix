{ pkgs }:

let
  script = pkgs.writeShellApplication {
    name = "emoji.sh";

    runtimeInputs = with pkgs; [
      fuzzel
      wl-clipboard
    ];

    text = ''
      set -eu

      # unfortunately we have in this directory two different file "formats":
      # 1. emojis with emoji first then description
      # 2. ascii artwork with description first, then artwork

      emojiAndDescription="$(cat ${./emoji-data/emojis.csv} ${./emoji-data/emoticons.csv} ${./emoji-data/ascii-art.txt} | fuzzel --dmenu --prompt='Emoji:')"

      # dirty hack: all emoji descriptions contain "<small>" substring
      if [[ $emojiAndDescription == *"<small>"* ]]; then
        # "<emoji> <description>" => "<emoji>"
        emoji="$(echo -n "$emojiAndDescription" | cut -d ' ' -f 1)"
      else
        # "<description> <tab> <ascii art>" => "<ascii art>"
        emoji="$(echo -n "$emojiAndDescription" | cut -d '	' -f 2)"
      fi

      # copy to clipboard
      echo -n "$emoji" | wl-copy
    '';
  };
  desktopItem = pkgs.makeDesktopItem {
    name = "emoji";
    desktopName = "Pick emoji";
    genericName = "Custom Script";
    exec = "${script}/bin/emoji.sh";
    icon = "application-x-executable";
    terminal = false;
    categories = [ "Utility" ];
  };
in
pkgs.symlinkJoin {
  name = "emoji";
  paths = [
    script
    desktopItem
  ];
}
