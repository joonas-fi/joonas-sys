{ modulesPath, pkgs, ... }:

{
  imports = [
    (modulesPath + "/installer/scan/not-detected.nix")
    ../../features/common.nix
    ../../features/developer-tools.nix
    ../../features/desktop.nix
    ./hardware-configuration.nix
    ./disk-config.nix
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
      intel-vaapi-driver # For older processors. LIBVA_DRIVER_NAME=i965
    ];
  };
  # environment.sessionVariables = { LIBVA_DRIVER_NAME = "iHD"; }; # Optionally, set the environment variable

  networking = {
    hostName = "old-laptop";
    networkmanager = {
      wifi.powersave = true;
    };
  };

  systemd.network.links."10-wifi" = {
    matchConfig.PermanentMACAddress = "4c:34:88:c2:82:ce";
    linkConfig.Name = "wifi";
  };

  environment.systemPackages = [
    pkgs.libva-utils
  ];

  services = {
    udev.extraRules = ''
      # cpu temp (Intel), temp1 = package temperature
      ACTION=="add", SUBSYSTEM=="hwmon", ATTR{name}=="coretemp", RUN+="/bin/sh -c 'ln -s /sys$devpath/temp1_input /dev/cpu_temp'"
    '';
  };
}
