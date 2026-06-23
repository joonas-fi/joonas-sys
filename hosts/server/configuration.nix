{ modulesPath, pkgs, ... }:

{
  imports = [
    (modulesPath + "/installer/scan/not-detected.nix")
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
      intel-media-driver
      intel-vaapi-driver # For older processors. LIBVA_DRIVER_NAME=i965
    ];
  };
  # environment.sessionVariables = { LIBVA_DRIVER_NAME = "iHD"; }; # Optionally, set the environment variable

  networking = {
    hostName = "server";
  };

  environment.systemPackages = [
    pkgs.libva-utils
  ];

  services = {
    udev.extraRules = ''
      # Raven Ridge [Radeon Vega Series / Radeon Vega Mobile Series]
      KERNEL=="renderD[0-9]*", SUBSYSTEM=="drm", ATTRS{vendor}=="0x1002", ATTRS{device}=="0x15dd", SYMLINK+="dri/amdgpu"
    '';
  };
}
