{ pkgs }:

let
  script = pkgs.writeShellApplication {
    name = "mirror-screen.sh";

    text = ''
      set -eu

      ${pkgs.wl-mirror}/bin/wl-mirror --scaling exact --scaling nearest --region "$(${pkgs.slurp}/bin/slurp -f '%x,%y %wx%h %o')"
    '';
  };
in
pkgs.makeDesktopItem {
  name = "mirror-screen";
  desktopName = "Mirror part of screen in a window";
  genericName = "Custom Script";
  exec = "${script}/bin/mirror-screen.sh";
  icon = "application-x-executable";
  terminal = false;
  categories = [ "Utility" ];
}
