{
  pkgs,
  lib,
  ...
}:

let
  varasto = pkgs.callPackage ../../packages/varasto.nix { };
in
{
  systemd.tmpfiles.rules = [
    "d /sto 0700 joonas users -" # mountpoint
  ];

  home-manager = {
    users."joonas" =
      { config, ... }:
      {
        home.file = {
          ".local/varasto-work/".source =
            config.lib.file.mkOutOfStoreSymlink "/sysroot/apps/varasto/varasto-work/";
          ".config/varasto/".source = config.lib.file.mkOutOfStoreSymlink "/sysroot/apps/varasto/";
        };
      };
  };

  # required by Varasto
  programs.fuse.userAllowOther = true;

  systemd.user.services.varasto-fs = {
    enable = true;

    description = "Varasto FUSE server for accessing files over network";

    wantedBy = [ "default.target" ];

    serviceConfig = {
      Type = "simple";
      ExecStart = "${varasto}/bin/sto fuse serve";
      Restart = "on-failure";
      RestartSec = "10s";
      Environment = "PATH=/run/wrappers/bin:%h/.nix-profile/bin:/run/current-system/sw/bin";
    };
  };

  environment.systemPackages = with pkgs; [
    varasto
  ];
}
