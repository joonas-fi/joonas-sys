{
  description = "All my system configurations";

  inputs = {
    # release announcements: https://discourse.nixos.org/c/announcements/nixos-releases/29
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    nixos-hardware.url = "github:NixOS/nixos-hardware";
    disko = {
      url = "github:nix-community/disko";
      # Use the same version as nixpkgs
      inputs.nixpkgs.follows = "nixpkgs";
    };
    home-manager = {
      # NOTE: home-manager version needs to match NixOS version. I tried newer here and it errored.
      url = "github:nix-community/home-manager/release-26.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    microvm = {
      url = "github:microvm-nix/microvm.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      nixpkgs,
      nixos-hardware,
      disko,
      home-manager,
      microvm,
      ...
    }:
    let
      bedroom-pi = nixpkgs.lib.nixosSystem {
        system = "aarch64-linux";
        modules = [
          "${nixpkgs}/nixos/modules/installer/sd-card/sd-image-aarch64.nix"
          nixos-hardware.nixosModules.raspberry-pi-4
          ./hosts/bedroom-pi/configuration.nix
        ];
      };

      devVM = import ./hosts/dev-vm/configuration.nix;

      mkDevVm =
        devVmSlot:
        nixpkgs.lib.nixosSystem {
          system = "x86_64-linux";
          specialArgs = {
            inherit devVmSlot;
          };
          modules = [
            microvm.nixosModules.microvm
            home-manager.nixosModules.default
            devVM.module
          ];
        };
    in
    {
      nixosConfigurations = {
        bedroom-pi = bedroom-pi;

        garage = nixpkgs.lib.nixosSystem {
          system = "aarch64-linux";
          modules = [
            "${nixpkgs}/nixos/modules/installer/sd-card/sd-image-aarch64.nix"
            nixos-hardware.nixosModules.raspberry-pi-3
            ./hosts/garage/configuration.nix
          ];
        };

        mediaplayer = nixpkgs.lib.nixosSystem {
          system = "aarch64-linux";
          modules = [
            "${nixpkgs}/nixos/modules/installer/sd-card/sd-image-aarch64.nix"
            home-manager.nixosModules.default # this is logic (not content) for processing home-manager directives
            nixos-hardware.nixosModules.raspberry-pi-5
            ./hosts/mediaplayer/configuration.nix
          ];
        };

        old-laptop = nixpkgs.lib.nixosSystem {
          system = "x86_64-linux";
          modules = [
            disko.nixosModules.disko # this is logic (not content) for processing disko directives
            home-manager.nixosModules.default # this is logic (not content) for processing home-manager directives
            ./hosts/old-laptop/configuration.nix
          ];
        };

        server = nixpkgs.lib.nixosSystem {
          system = "x86_64-linux";
          modules = [
            disko.nixosModules.disko # this is logic (not content) for processing disko directives
            home-manager.nixosModules.default # this is logic (not content) for processing home-manager directives
            ./hosts/server/configuration.nix
          ];
        };

        work = nixpkgs.lib.nixosSystem {
          system = "x86_64-linux";
          modules = [
            # disko.nixosModules.disko # this is logic (not content) for processing disko directives
            home-manager.nixosModules.default # this is logic (not content) for processing home-manager directives
            ./hosts/work/configuration.nix
          ];
        };
      }
      // builtins.listToAttrs (
        map (devVmSlot: {
          name = devVmSlot.name;
          value = mkDevVm devVmSlot;
        }) devVM.slots
      );

      packages.aarch64-linux = {
        bedroom-pi-sd-image = bedroom-pi.config.system.build.sdImage;
      };

      # callable with `$ nix fmt`
      formatter.x86_64-linux = nixpkgs.legacyPackages.x86_64-linux.nixfmt-tree;
    };
}
