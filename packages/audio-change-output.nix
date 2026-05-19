{ pkgs }:

let
  script = pkgs.writeShellApplication {
    name = "audio-change-output.sh";

    runtimeInputs = with pkgs; [
      fuzzel
    ];

    text = ''
      set -eu

      sink="$(pactl list short sinks | cut -f 2 | fuzzel --dmenu --prompt='Audio output:')";

      pactl set-default-sink "$sink"
    '';
  };
  desktopItem = pkgs.makeDesktopItem {
    name = "audio-change-output";
    desktopName = "Audio change output";
    genericName = "Custom Script";
    exec = "${script}/bin/audio-change-output.sh";
    icon = "application-x-executable";
    terminal = false;
    categories = [ "Utility" ];
  };
in
pkgs.symlinkJoin {
  name = "audio-change-output";
  paths = [
    script
    desktopItem
  ];
}
