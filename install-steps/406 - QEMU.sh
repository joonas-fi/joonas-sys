#!/bin/bash

source common.sh


# Virtual Machines
# also uses Samba (for VM shared folders), which we install earlier
# qemu-system-misc for RISC-V and some other exotic architectures.
# qemu-efi-aarch64 to have EFI firmware

apt install -y qemu-system-x86 qemu-system-arm qemu-system-misc \
	qemu-efi-aarch64

#versioncommand: qemu-system-x86_64 --version
