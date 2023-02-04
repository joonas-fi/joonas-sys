#!/bin/bash

source common.sh


# Virtual Machines
# also uses Samba (for VM shared folders), which we install earlier
# qemu-system-misc for RISC-V and some other exotic architectures.

apt install -y qemu-system-x86 qemu-system-misc

#versioncommand: qemu-system-x86_64 --version
