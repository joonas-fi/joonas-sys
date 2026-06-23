{ lib, pkgs, ... }:

{
  hardware = {
    raspberry-pi."4" = {
      apply-overlays-dtmerge.enable = true;
      fkms-3d.enable = true;
    };
    deviceTree = {
      enable = true;
      # filter = "*rpi-4-*.dtb";
    };
    graphics.enable = true;
  };

  # overcome some bug in nixos-hardware/raspberry-pi/4
  # https://discourse.nixos.org/t/cannot-build-raspberry-pi-sdimage-module-dw-hdmi-not-found/71804
  boot.initrd.allowMissingModules = true;

  nixpkgs.hostPlatform = "aarch64-linux";

  image.baseName = "nixos-bedroom-pi";

  networking = {
    hostName = "bedroom-pi";
    networkmanager = {
      wifi.powersave = true;
    };
  };

  users.users.joonas = {
    extraGroups = [
      # already come from common
      # "input" # moonlight to use input devices in console mode
      # "video" # needed for `/dev/dri/...` access (moonlight uses direct rendering)
      # "render"
    ];
  };

  environment.systemPackages = with pkgs; [
    bluez # for bluetoothctl

    vim
    moonlight-qt
    libraspberrypi
    raspberrypi-eeprom
    v4l-utils
    ffmpeg
  ];
}
