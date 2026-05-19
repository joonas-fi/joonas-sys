{ pkgs }:

pkgs.runCommand "my-wallpaper" { } ''
  mkdir -p $out/share/backgrounds
  cp ${pkgs.pop-wallpapers}/share/backgrounds/pop/benjamin-voros-250200.jpg \
     $out/share/backgrounds/wallpaper.jpg
''
