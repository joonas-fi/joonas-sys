{ modulesPath, pkgs, ... }:

{
  imports = [
    (modulesPath + "/installer/scan/not-detected.nix")
    ./hardware-configuration.nix
    ./disk-config.nix
  ];

  boot.loader.grub.enable = true;

  networking = {
    hostName = "moist-helen";
    domain = "fn61.net";

    firewall = {
      enable = true;
      # edgerouter (deployed out-of-band)
      allowedTCPPorts = [
        80
        443
      ];
    };
  };
}
