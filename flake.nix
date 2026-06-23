{
  description = "All my system configurations";

  inputs = {
    # release announcements: https://discourse.nixos.org/c/announcements/nixos-releases/29
    # most of the system comes from here
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    # for the few things that need to move faster than the stable released version
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
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
      self,
      nixpkgs,
      nixpkgs-unstable,
      nixos-hardware,
      disko,
      home-manager,
      microvm,
      ...
    }:
    let
      devVM = import ./hosts/dev-vm/configuration.nix;
      generateConfigurationsDocumentation = import ./misc/generate-configurations-documentation.nix;

      # list is hardcoded here (as opposed to listing from filesystem) to have a logical column order.
      featurePaths = map (feature: ./features + "/${feature}.nix") [
        "common"
        "common-extended"
        "developer-tools"
        "desktop"
      ];

      # this indirection required for introspection (see "generate configurations documentation")
      mkNixosSystem =
        args:
        (nixpkgs.lib.nixosSystem args)
        // {
          architecture = nixpkgs.lib.removeSuffix "-linux" args.system;
          featureImports = builtins.filter (
            module: builtins.isPath module && builtins.elem (toString module) (map toString featurePaths)
          ) args.modules;
        };

      mkDevVm =
        devVmSlot:
        mkNixosSystem {
          system = "x86_64-linux";
          specialArgs = {
            inherit devVmSlot;
          };
          modules = [
            microvm.nixosModules.microvm
            home-manager.nixosModules.default
            devVM.module
            ./features/common.nix
            ./features/common-extended.nix
            ./features/developer-tools.nix
          ];
          specialArgs = {
            # pass this as input for the modules
            pkgsUnstable = nixpkgs-unstable.legacyPackages.x86_64-linux;
          };
        };
    in
    {
      nixosConfigurations = {
        bedroom-pi = mkNixosSystem {
          system = "aarch64-linux";
          modules = [
            "${nixpkgs}/nixos/modules/installer/sd-card/sd-image-aarch64.nix"
            nixos-hardware.nixosModules.raspberry-pi-4
            home-manager.nixosModules.default
            ./hosts/bedroom-pi/configuration.nix
            ./features/common.nix
          ];
        };

        garage = mkNixosSystem {
          system = "aarch64-linux";
          modules = [
            "${nixpkgs}/nixos/modules/installer/sd-card/sd-image-aarch64.nix"
            nixos-hardware.nixosModules.raspberry-pi-3
            ./hosts/garage/configuration.nix
          ];
        };

        mediaplayer = mkNixosSystem {
          system = "aarch64-linux";
          modules = [
            "${nixpkgs}/nixos/modules/installer/sd-card/sd-image-aarch64.nix"
            home-manager.nixosModules.default
            nixos-hardware.nixosModules.raspberry-pi-5
            ./hosts/mediaplayer/configuration.nix
          ];
        };

        old-laptop = mkNixosSystem {
          system = "x86_64-linux";
          modules = [
            disko.nixosModules.disko # this is logic (not content) for processing disko directives
            home-manager.nixosModules.default # this is logic (not content) for processing home-manager directives
            ./hosts/old-laptop/configuration.nix
            ./features/common.nix
            ./features/common-extended.nix
            ./features/developer-tools.nix
            ./features/desktop.nix
          ];
          specialArgs = {
            # pass this as input for the modules
            pkgsUnstable = nixpkgs-unstable.legacyPackages.x86_64-linux;
          };
        };

        workworkwork = mkNixosSystem {
          system = "x86_64-linux";
          modules = [
            home-manager.nixosModules.default
            ./hosts/workworkwork/configuration.nix
            ./features/common.nix
            ./features/common-extended.nix
            ./features/developer-tools.nix
            ./features/desktop.nix
            ./features/sub-features/microvm-host.nix
          ];
          specialArgs = {
            # pass this as input for the modules
            pkgsUnstable = nixpkgs-unstable.legacyPackages.x86_64-linux;
          };
        };

        server = mkNixosSystem {
          system = "x86_64-linux";
          modules = [
            disko.nixosModules.disko
            home-manager.nixosModules.default
            ./hosts/server/configuration.nix
            ./features/common.nix
            ./features/common-extended.nix
            ./features/developer-tools.nix
          ];
        };

        "moist-helen.fn61.net" = mkNixosSystem {
          system = "x86_64-linux";
          modules = [
            disko.nixosModules.disko
            home-manager.nixosModules.default
            ./hosts/moist-helen.fn61.net/configuration.nix
            ./features/common.nix
            ./features/common-extended.nix
          ];
        };

        work = mkNixosSystem {
          system = "x86_64-linux";
          modules = [
            # disko.nixosModules.disko
            home-manager.nixosModules.default
            ./hosts/work/configuration.nix
            ./features/common.nix
            ./features/common-extended.nix
            ./features/developer-tools.nix
            ./features/desktop.nix
            ./features/sub-features/microvm-host.nix
          ];
          specialArgs = {
            # pass this as input for the modules
            pkgsUnstable = nixpkgs-unstable.legacyPackages.x86_64-linux;
          };
        };
      }
      // builtins.listToAttrs (
        map (devVmSlot: {
          name = devVmSlot.name;
          value = mkDevVm devVmSlot;
        }) devVM.slots
      );

      configurationsMarkdown = generateConfigurationsDocumentation.markdown self.nixosConfigurations featurePaths;

      apps.x86_64-linux.generate-configurations-documentation = {
        type = "app";
        program = "${
          nixpkgs.legacyPackages.x86_64-linux.writeShellApplication {
            name = "generate-configurations-documentation";
            runtimeInputs = [ nixpkgs.legacyPackages.x86_64-linux.nix ];
            text = ''
              nix eval --raw .#configurationsMarkdown > configurations.md
            '';
          }
        }/bin/generate-configurations-documentation";
      };

      # callable with `$ nix fmt`
      formatter.x86_64-linux = nixpkgs.legacyPackages.x86_64-linux.nixfmt-tree;
    };
}
