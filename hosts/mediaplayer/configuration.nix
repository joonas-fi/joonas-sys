{ lib, pkgs, ... }:

{
  imports = [
    ../../features/common.nix
    ../../features/desktop.nix
  ];

  hardware = {
    # raspberry-pi."5" = {
    #   apply-overlays-dtmerge.enable = true;
    #   fkms-3d.enable = true;
    # };
    deviceTree = {
      enable = true;
      # filter = "*rpi-4-*.dtb";
    };
  };

  # overcome some bug in nixos-hardware/raspberry-pi/4
  # https://discourse.nixos.org/t/cannot-build-raspberry-pi-sdimage-module-dw-hdmi-not-found/71804
  boot.initrd.allowMissingModules = true;

  image.baseName = "nixos-mediaplayer";

  networking = {
    hostName = "mediaplayer";
  };

  system.stateVersion = "25.11";
}
