{ lib, pkgs, ... }:

{
  hardware = {
    # raspberry-pi."3" = {
    #   apply-overlays-dtmerge.enable = true;
    #   fkms-3d.enable = true;
    # };
    deviceTree = {
      enable = true;
      # filter = "*rpi-4-*.dtb";
    };
    graphics.enable = true;
    enableRedistributableFirmware = true;
  };

  # overcome some bug in nixos-hardware/raspberry-pi/4
  # https://discourse.nixos.org/t/cannot-build-raspberry-pi-sdimage-module-dw-hdmi-not-found/71804
  boot.initrd.allowMissingModules = true;

  nixpkgs.hostPlatform = "aarch64-linux";

  image.baseName = "nixos-garage";

  networking = {
    hostName = "garage";
    networkmanager = {
      enable = true;
      wifi.powersave = true;
    };
  };

  services = {
    tailscale = {
      enable = true;
      openFirewall = true;
    };

    openssh = {
      enable = true;
      openFirewall = true;
      settings = {
        # PasswordAuthentication = true;
        PermitRootLogin = "no";
      };
    };

    avahi = {
      enable = true;
      nssmdns4 = true;
      openFirewall = true;
    };
  };

  users.users.joonas = {
    isNormalUser = true;
    # initialPassword = "nixos";
    extraGroups = [
      "networkmanager"
      "wheel"
      "input" # moonlight to use input devices in console mode
      "video" # needed for `/dev/dri/...` access (moonlight uses direct rendering)
      "render"
    ];
    openssh.authorizedKeys.keys = [
      "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAID0kw624nwuYMt0rju0MRfkvT6Opmww52kmlnQC71LtB joonas@work"
    ];
  };

  programs.sway = {
    enable = true;
    wrapperFeatures.gtk = true;
  };

  environment.systemPackages = with pkgs; [
    # git
    vim
    moonlight-qt
    libraspberrypi
    raspberrypi-eeprom
    v4l-utils
    ffmpeg
  ];

  # don't ask for passwords. this is also needed by nixos-rebuild switch to not need asking sudo passwords.
  security.sudo.wheelNeedsPassword = false;

  nix.settings = {
    trusted-users = [
      "root"
      "@wheel" # otherwise nixos-rebuild switch gives `cannot add path ... because it lacks a signature by a trusted key` errors
    ];
  };

  system.stateVersion = "25.11";
}
