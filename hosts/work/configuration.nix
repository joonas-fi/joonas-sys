{ modulesPath, pkgs, ... }:

{
  imports = [
    (modulesPath + "/installer/scan/not-detected.nix")
    ../../features/common.nix
    ../../features/developer-tools.nix
    ../../features/desktop.nix
    ../../features/sub-features/microvm-host.nix
    ./hardware-configuration.nix
    ./sub-features/unikie.nix
  ];

  boot = {
    loader.systemd-boot = {
      enable = true;
      configurationLimit = 10;
    };
  };

  networking = {
    hostName = "work";
  };

  systemd.network.links."10-wired" = {
    matchConfig.PermanentMACAddress = "18:31:bf:30:5c:b1";
    linkConfig.Name = "wired";
  };

  services = {
    udev.extraRules = ''
      # AMD Ryzen 9 5900X (Matisse Device 24: Function 3)
      # temp3 is `Tccd1` (temp1 would be `Tctl`)
      ACTION=="add", SUBSYSTEM=="hwmon", ATTRS{vendor}=="0x1022", ATTRS{device}=="0x1443", RUN+="/bin/sh -c 'ln -s /sys$devpath/temp3_input /dev/cpu_temp'"
    '';
  };
}
