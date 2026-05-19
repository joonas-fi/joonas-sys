{ lib, pkgs, ... }:
{
  # make bash available to programs hardcoding the path `/bin/bash`
  services.envfs.enable = true;

  # needs to be as a service
  services.udisks2.enable = true;

  environment = {
    # tii-yocto flash Nvidia deps
    systemPackages = with pkgs; [
      bmaptool
      dtc
      gcc
      gnumake
      gptfdisk
      parted # for partprobe
      (python3.withPackages (
        ps: with ps; [
          pyyaml
        ]
      ))

      picocom # for connecting to the device
    ];
  };

  # tii Tailscale <-> GlobalConnect proxy
  networking.extraHosts = ''
    10.151.12.79  confluence.tii.ae
    10.151.12.100  i-service.tii.ae
  '';

  # remote builder doesn't work without the host key being in root's SSH known hosts
  programs.ssh.knownHosts."builder.vedenemo.dev" = {
    publicKey = "AAAAC3NzaC1lZDI1NTE5AAAAIG68NdmOw3mhiBZwDv81dXitePoc1w//p/LpsHHA8QRp";
  };

  nix = {
    distributedBuilds = true;

    settings.builders-use-substitutes = true;

    buildMachines = [
      {
        hostName = "builder.vedenemo.dev";
        system = "aarch64-linux";

        protocol = "ssh-ng";
        sshUser = "joonas";
        sshKey = "/dev/shm/id_ed25519";

        maxJobs = 4;
        speedFactor = 1;

        supportedFeatures = [
          "big-parallel"
          "nixos-test"
          "kvm"
        ];
      }
    ];
  };
}
