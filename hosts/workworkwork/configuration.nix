{ modulesPath, pkgs, ... }:

{
  imports = [
    (modulesPath + "/installer/scan/not-detected.nix")
    ./hardware-configuration.nix
  ];

  boot = {
    loader.systemd-boot = {
      enable = true;
      configurationLimit = 10;
    };
  };

  hardware.graphics = {
    extraPackages = with pkgs; [
      # https://wiki.nixos.org/wiki/Accelerated_Video_Playback
      # intel-media-driver
    ];
  };

  networking = {
    hostName = "workworkwork";
    networkmanager = {
      # wifi.powersave = true;
    };
  };

  systemd.network.links."10-wifi" = {
    matchConfig.PermanentMACAddress = "dc:56:7b:6c:be:b9";
    linkConfig.Name = "wifi";
  };

  environment.systemPackages = [
  ];

  services = {
    # udev.extraRules = ''
    #   # Asus Vivobook (AMD Ryzen), temp1 = Tctl (no other options available)
    #   ACTION=="add", SUBSYSTEM=="hwmon", ATTR{name}=="k10temp", RUN+="/bin/sh -c 'ln -s /sys$devpath/temp1_input /dev/cpu_temp'"

    #   # AMD GPU backlight
    #   ACTION=="add", SUBSYSTEM=="backlight", KERNEL=="amdgpu_bl1", GROUP="joonas", MODE="0664", RUN+="/bin/chgrp users /sys$devpath/brightness", RUN+="/bin/chmod g+w /sys$devpath/brightness" RUN+="/bin/ln -s /sys$devpath /dev/screen-backlight"
    # '';
  };
}
