#!/bin/bash

source common.sh


# Virtual Machines
# also uses Samba (for VM shared folders), which we install earlier

apt install -y qemu-system-x86

#versioncommand: qemu-system-x86_64 --version
