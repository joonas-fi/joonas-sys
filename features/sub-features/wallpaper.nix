{ pkgs }:

pkgs.runCommand "my-wallpaper" { } ''
  mkdir -p $out/share/backgrounds
  # pick one from here: https://github.com/pop-os/wallpapers/tree/master/original
  # cp ${pkgs.pop-wallpapers}/share/backgrounds/pop/jad-limcaco-183877.jpg \
  cp ${pkgs.pop-wallpapers}/share/backgrounds/pop/jasper-van-der-meij-97274-edit.jpg \
     $out/share/backgrounds/wallpaper.jpg
''
