{ pkgs, config, ... }:
{
  imports = [
    ./sub-features/helix.nix
  ];

  # NOTE: virtualization is spelled wrong 😬
  virtualisation = {
    docker.enable = true;
    containerd.enable = true; # looks to be maybe required?
  };

  users.users.joonas = {
    extraGroups = [
      "docker"
    ];
  };
}
