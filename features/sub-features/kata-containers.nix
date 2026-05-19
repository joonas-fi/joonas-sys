{
  pkgs,
  ...
}:

{
  # Make containerd able to find containerd-shim-kata-v2 in the Nix store.
  systemd.services.containerd.path = with pkgs; [
    kata-runtime
  ];

  environment.systemPackages = with pkgs; [
    kata-runtime
    containerd
    nerdctl
  ];
}
