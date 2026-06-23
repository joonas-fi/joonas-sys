{ modulesPath, pkgs, ... }:

{
  imports = [
    (modulesPath + "/installer/scan/not-detected.nix")
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

      # logitech MX ergo
      # we cannot use the symlink facility for anything that doesn't end up in /dev?
      ACTION=="add|change", SUBSYSTEM=="power_supply", ATTR{manufacturer}=="Logitech", ATTR{model_name}=="MX Ergo Multi-Device Trackball ", RUN+="/bin/sh -c 'ln -s /sys$devpath /dev/mouse-battery'"
      # ACTION=="add|change",    SUBSYSTEM=="power_supply", ATTR{manufacturer}=="Logitech", ATTR{model_name}=="MX Ergo Multi-Device Trackball ", RUN+="/bin/sh -c 'ln -s /sys$devpath /dev/mouse-battery'"
      ACTION=="remove", SUBSYSTEM=="power_supply", ATTR{manufacturer}=="Logitech", ATTR{model_name}=="MX Ergo Multi-Device Trackball ",                 RUN+="/bin/sh -c 'rm /dev/mouse-battery'"
    '';
  };

  # the device advertises it can do hw volume control but it doesn't work (0 % vol still is effectively full blast), so we force soft mixer
  services.pipewire.wireplumber.extraConfig."99-br21-soft-mixer" = {
    "monitor.alsa.rules" = [
      {
        matches = [
          {
            # https://www.amazon.de/dp/B09KN343S4
            # regex to allow for `.2` etc suffixes if device re-connected
            "device.name" = "~^alsa_card\\.usb-ZhuHai_JieLi_Technology_JieLi_BR21_20180105-01.*$";
          }
        ];
        actions.update-props = {
          "api.alsa.soft-mixer" = true;
        };
      }
    ];
  };
}
