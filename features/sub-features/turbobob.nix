{
  pkgs,
  lib,
  ...
}:

let
  turbobob = pkgs.callPackage ../../packages/turbobob.nix { };
in
{
  systemd.services.turbobob-proxy = {
    enable = true;

    description = "Edgerouter embeddable proxy";

    wantedBy = [ "default.target" ];

    serviceConfig = {
      Type = "simple";
      ExecStart = "${turbobob}/bin/bob tools proxy serve";
      Restart = "on-failure";
      RestartSec = "10s";
      Environment = [
        "DOCKER_URL=unix:///var/run/docker.sock" # discover (HTTP) apps automatically through Docker
        "NETWORK_NAME=docker_gwbridge" # discover apps' IP addresses from perspective of this network
      ];
    };
  };

  environment.systemPackages = with pkgs; [
    turbobob
  ];
}
