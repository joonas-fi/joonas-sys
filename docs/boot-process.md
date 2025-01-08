Boot process
============

The boot partition is an [EFI System Partition](https://en.wikipedia.org/wiki/EFI_system_partition).
It contains EFI-bootable [Unified Kernel Image (UKI)](https://wiki.archlinux.org/title/Unified_kernel_image) which
combines into one EFI-bootable file containing:

- the Linux kernel
- [EFI boot stub](https://wiki.archlinux.org/title/EFI_boot_stub) (aka EFI stub)
	* this "adapts" the Linux kernel to boot when invoked by EFI firmware (which expects EFI executable).
- [kernel commandline](https://docs.kernel.org/admin-guide/kernel-parameters.html) and
- [initrd](https://en.wikipedia.org/wiki/Initial_ramdisk).

This means we don't need traditional additional bootloader like Grub.
The commandline instructs which checkout to use.

```mermaid
flowchart TD
    subgraph "Firmware"
        boot(Device boot) --> uefiboot
    uefiboot[Firmware bootloader
    UEFI]
    end
    subgraph "ESP"
        osbootloader(run:
    /EFI/BOOT/BOOTx64.efi
    UKI as EFI app)
        efistub(EFI Stub
    - outer interface: EFI
    - inner interface: Linux loader
    - passes embedded resources to Linux kernel)
        bootLinux(Boot Linux kernel)
        Initrd2[Initrd
    - mount root partition to /sysroot
    - mount checkout as / overlay based on cmdline]
    end
    subgraph "UKI (single-file)"
        kernel[Kernel + EFI stub] --> Resources(Embedded resources)
        Initrd(Initrd
        - 'early userspace'
        - knows how to find & mount root partition) --> Resources
        Cmdline --> Resources
    end
    uefiboot -->|Find ESP| osbootloader
    osbootloader --> efistub
    efistub --> bootLinux
    bootLinux --> Initrd2
    Initrd -.- Initrd2
    Initrd2 --> userspaceboot[Userspace boot]
    Resources -. contained in .-> osbootloader
```