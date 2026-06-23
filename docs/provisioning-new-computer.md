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


Raspberry Pi
------------

### Local building or remote building?

The often documented way to work with Raspberry Pi is to flash official NixOS image and then start editing your NixOS config on the end device.

This is unacceptable as Raspberry Pi is so light on CPU / RAM / SD card horsepower that the only sensible strategy is to
build on a more powerful PC and then deploy the already-built artifacts to the Raspberry Pi over network.


### How to bootstrap?

Bootstrap = get first version of your NixOS config on the device, after which it's easier to deploy changes.

One could:

- option 1: boot Raspberry Pi with an official NixOS image and then use nixos-anywhere to deploy your own configuration from another machine with `nixos-rebuild switch --target-host=...`

- option 2: or we can take configuration that we build in our build machine, build the config as SD image and flash that
  * then changes after that are `nixos-rebuild switch --target-host=...`

PREREQUISITE: have your config import module `${nixpkgs}/nixos/modules/installer/sd-card/sd-image-aarch64.nix`

We're documenting option 2 here.

Suppose you have your standard build command for a flake-based config:

```shell
nixos-rebuild build --flake=.#example-pi
```

Sidenote: the bare-nix form for this is:

```shell
nix build .#nixosConfigurations.example-pi.config.system.build.toplevel
```

Now, to build the SD image of that config we replace `toplevel` with `sdImage`. The final build command is:

```shell
nix build .#nixosConfigurations.example-pi.config.system.build.sdImage
```

The `result` symlink in your workdir points to directory where there is zstd-compressed file `sd-image/nixos-example-pi.img.zst`

You can on-the-fly decompress and flash that to your SD card:

```shell
cat result/sd-image/nixos-example-pi.img.zst  | zstd -d | sudo dd of=/dev/... bs=1M status=progress
```


Hetzner install example
-----------------------

Follow the outline of installing to new computer. The NixOS-anywhere works when provisioned on top of the regular old Ubuntu..

[Some more docs](https://wiki.nixos.org/wiki/Install_NixOS_on_Hetzner_Cloud)

Flake:

```nix
        "moist-helen.fn61.net" = nixpkgs.lib.nixosSystem {
          system = "x86_64-linux";
          modules = [
            disko.nixosModules.disko # this is logic (not content) for processing disko directives
            home-manager.nixosModules.default # this is logic (not content) for processing home-manager directives
            ./hosts/moist-helen.fn61.net/configuration.nix
          ];
        };
```

Write `hosts/moist-helen.fn61.net/configuration.nix`:

```nix
{ modulesPath, pkgs, ... }:

{
  imports = [
    (modulesPath + "/installer/scan/not-detected.nix")
    ../../features/common.nix
    ./hardware-configuration.nix
    ./disk-config.nix
  ];

  boot.loader.grub.enable = true;

  networking = {
    hostName = "moist-helen";
    domain = "fn61.net";
  };
}
```

Write `hosts/moist-helen.fn61.net/disk-config.nix`:

```nix
{ lib, ... }:

{
  disko.devices = {
    disk= {
    main = {
      type = "disk";
      device = "/dev/sda";
      content = {
        type = "gpt";
        partitions = {
          boot = {
            size = "1M";
            type = "EF02";
            priority = 1;
          };
          ESP = {
            size = "512M";
            type = "EF00";
            content = {
              type = "filesystem";
              format = "vfat";
              mountpoint = "/boot";
            };
          };
          root = {
            size = "100%";
            content = {
              type = "filesystem";
              format = "ext4";
              mountpoint = "/";
            };
          };
        };
      };
      };
    };
  };
}
```


