Provisioning new computer
=========================

From blank slate
----------------

1. Add `hosts/<computer>/configuration.nix`
  * remember to have `initialPassword` because you can't access computer over SSH on first boot due to not connected to wifi yet
1. Add `hosts/<computer>/disk-config.nix`
1. Do not worry about not having `hardware-configuration.nix` yet as the `nixos-anywhere` command will create it
1. USB-boot computer with [NixOS minimal ISO image](https://nixos.org/download/#minimal-iso-image), connect to wifi with `$ nmtui`
1. From another computer run `nix run github:nix-community/nixos-anywhere -- --flake .#old-laptop --generate-hardware-config nixos-generate-config hosts/<computer>/hardware-configuration.nix --target-host nixos@<your IP>`

Lots of inspiration from [Michael Stapelberg](https://michael.stapelberg.ch/posts/2025-06-01-nixos-installation-declarative/) - thanks!


From existing system
--------------------

```shell
nix shell nixpkgs#nixos-install-tools nixpkgs#util-linux
nixos-generate-config  # move this `hardware-configuration.nix` into this repo proper place
system=$(nix build --print-out-paths .#nixosConfigurations.work.config.system.build.toplevel)
nixos-install --root /workspace/target --system "$system"
```

NOTE: if it gives you shit about mount command try this:

```shell
env "PATH=$PATH:/nix/var/nix/profiles/system/sw/bin:/usr/sbin:/sbin"   "$(command -v nixos-install)" --flake .#work --root /sysroot
```
