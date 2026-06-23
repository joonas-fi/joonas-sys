{ lib, pkgs }:

let
  mkScript =
    name: file:
    pkgs.writeShellApplication {
      inherit name;
      text = builtins.readFile file;
    };

  ffmpeg-concat-video = mkScript "ffmpeg-concat-video.sh" ./ffmpeg-joonas-scripts/ffmpeg-concat-video.sh;
  ffmpeg-cut-video = mkScript "ffmpeg-cut-video.sh" ./ffmpeg-joonas-scripts/ffmpeg-cut-video.sh;
  ffmpeg-transmux = mkScript "ffmpeg-transmux.sh" ./ffmpeg-joonas-scripts/ffmpeg-transmux.sh;
in
pkgs.symlinkJoin {
  name = "ffmpeg-joonas-scripts";

  paths = [
    ffmpeg-concat-video
    ffmpeg-cut-video
    ffmpeg-transmux
  ];

  meta = {
    description = "Joonas' ffmpeg helper scripts";
    platforms = lib.platforms.unix;
  };
}
