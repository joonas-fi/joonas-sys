# thanks https://michael.stapelberg.ch/posts/2026-02-01-coding-agent-microvm-nix/
{ pkgs, config, ... }:
let
  bridgeName = "sandbox-br";
  dev-vm = pkgs.callPackage ../../packages/dev-vm.nix { };
in
{
  # since we are using systemd-networkd's features we've to enable it.
  # NOTE: be careful about systemd-network and networkmanager stomping on each other.
  systemd.network.enable = true;

  # create bridge
  systemd.network.netdevs."20-sandbox-br".netdevConfig = {
    Kind = "bridge";
    Name = bridgeName;
  };

  # configure bridge
  systemd.network.networks."20-sandbox-br" = {
    matchConfig.Name = bridgeName;
    addresses = [ { Address = "192.168.200.1/24"; } ];
    # do not try to wait for this network to become "online"
    linkConfig.RequiredForOnline = "no";
    networkConfig = {
      # configure even when there's no carrier (uplink/downlink to bridge attached?)
      ConfigureWithoutCarrier = true;
    };
  };

  # when new links are created (during VM launch) for microVM use, attach them to our bridge
  systemd.network.networks."21-microvm-tap" = {
    matchConfig.Name = "microvm*";
    networkConfig.Bridge = bridgeName;
  };

  # let the traffic flow between VMs and external network
  networking.nat = {
    enable = true;
    internalInterfaces = [ bridgeName ];
    externalInterface = "wired";
  };

  environment.systemPackages = [
    dev-vm
  ];
}
