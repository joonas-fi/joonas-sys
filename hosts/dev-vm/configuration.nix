let
  resources = {
    vcpu = 12; # 24 cores at host so figure two dev-vms going full blast is good allocation
    mem = 16 * 1024;
  };

  slots = [
    {
      name = "dev-vm-1";
      ip = "192.168.200.11";
      tap = "microvm-dev1";
      mac = "02:00:00:00:20:01";
      vsockCid = 11;
      machineId = "10000000-0000-4000-8000-000000000001";
    }
    {
      name = "dev-vm-2";
      ip = "192.168.200.12";
      tap = "microvm-dev2";
      mac = "02:00:00:00:20:02";
      vsockCid = 12;
      machineId = "10000000-0000-4000-8000-000000000002";
    }
    {
      name = "dev-vm-3";
      ip = "192.168.200.13";
      tap = "microvm-dev3";
      mac = "02:00:00:00:20:03";
      vsockCid = 13;
      machineId = "10000000-0000-4000-8000-000000000003";
    }
    {
      name = "dev-vm-4";
      ip = "192.168.200.14";
      tap = "microvm-dev4";
      mac = "02:00:00:00:20:04";
      vsockCid = 14;
      machineId = "10000000-0000-4000-8000-000000000004";
    }
    {
      name = "dev-vm-5";
      ip = "192.168.200.15";
      tap = "microvm-dev5";
      mac = "02:00:00:00:20:05";
      vsockCid = 15;
      machineId = "10000000-0000-4000-8000-000000000005";
    }
  ];

  module =
    {
      config,
      lib,
      pkgs,
      devVmSlot,
      ...
    }:

    let
      vmName = devVmSlot.name;
      bridgeInterface = "sandbox-br";
      slotRoot = "/var/lib/dev-vms/slots/${vmName}/root";
      ip = "${pkgs.iproute2}/bin/ip";
      machineId = lib.replaceStrings [ "-" ] [ "" ] devVmSlot.machineId;
    in
    {
      imports = [
        ../../features/common.nix
        ../../features/developer-tools.nix
      ];

      boot = {
        loader.grub.enable = false;

        initrd.kernelModules = [
          "fuse"
          "virtio_pci"
          "virtio_rng"
          "virtio_net"
          "virtiofs"
        ];

        supportedFilesystems = [ "virtiofs" ];
      };

      fileSystems."/" = lib.mkForce {
        device = "root";
        fsType = "virtiofs";
        neededForBoot = true;
        options = [
          "defaults"
          # The root virtio-fs mount needs the virtiofs module before systemd
          # attempts the mount during early boot.
          "x-systemd.after=systemd-modules-load.service"
        ];
      };

      networking = {
        hostName = vmName;
        # Keep the 32-bit host ID stable per slot, not per mounted project.
        hostId = lib.mkDefault (builtins.substring 0 8 machineId);
        useDHCP = false;
        useNetworkd = true;
        nameservers = [ "1.1.1.1" ];
      };

      systemd.network.networks."10-lan" = {
        matchConfig.Name = "en* eth*";
        networkConfig = {
          Address = "${devVmSlot.ip}/24";
          Gateway = "192.168.200.1";
          DNS = "1.1.1.1";
        };
      };

      services.resolved.enable = true;

      environment.etc."machine-id".text = "${machineId}\n";

      microvm = {
        hypervisor = "qemu";
        # The launcher starts the declared runner directly, instead of using
        # microvm.nix's host-side VM service integration.
        guest.enable = false;
        storeOnDisk = false;
        # Register this system closure in the guest Nix DB; the writable store
        # upperdir is slot state, so registrations should follow the slot.
        # without this we get `Failed to start Find NixOS closure` error.
        registerClosure = true;
        machineId = devVmSlot.machineId;
        vsock.cid = devVmSlot.vsockCid;
        vcpu = resources.vcpu;
        mem = resources.mem;

        # Boot QEMU directly into the selected NixOS system closure.
        kernelParams = config.boot.kernelParams ++ [
          "init=${config.system.build.toplevel}/init"
        ];

        virtiofsd.extraArgs = [
          "--announce-submounts" # The exported root contains overlayfs and bind-mount submounts.
          "--modcaps=+sys_admin" # Docker in VM uses overlayfs which likes to set `trusted.overlay.*` xattrs which requires `CAP_SYS_ADMIN` at the host side
        ];

        shares = [
          {
            tag = "root";
            socket = "/run/microvm-${vmName}-root.sock";
            source = slotRoot;
            mountPoint = "/";
            proto = "virtiofs";
            readOnly = false;
          }
        ];

        interfaces = [
          {
            type = "tap";
            id = devVmSlot.tap;
            mac = devVmSlot.mac;
            tap.vhost = true;
          }
        ];

        preStart = ''
          set -euo pipefail

          ${ip} link delete ${devVmSlot.tap} >/dev/null 2>&1 || true
          ${ip} tuntap add dev ${devVmSlot.tap} mode tap multi_queue
          ${ip} link set dev ${devVmSlot.tap} master ${bridgeInterface}
          ${ip} link set dev ${devVmSlot.tap} up
        '';
      };

      home-manager = {
        users."joonas" = {
          programs.opencode = {
            enable = true;

            # https://opencode.ai/docs/config/
            settings = {
              autoupdate = false;

              model = "openai/gpt-5.6-sol";
              small_model = "openai/gpt-5.6-terra";

              # try allow full permissions. (we're running in full VM sandbox)
              permission = "allow";
            };

            tui = {
              theme = "one-dark";
            };
          };

          programs.codex = {
            enable = true;
            settings = {
              default_permissions = ":danger-full-access";
            };
          };
        };
      };

      services = {
        # Enable qemu-guest-agent for standard QEMU guest integration.
        qemuGuest.enable = true;
      };

      systemd = {
        # Developer shells run Nix commands inside the guest.
        services.nix-daemon.enable = true;
        # Keep socket activation available for the Nix daemon.
        sockets.nix-daemon.enable = true;
      };
    };
in
{
  inherit
    module
    resources
    slots
    ;
}
