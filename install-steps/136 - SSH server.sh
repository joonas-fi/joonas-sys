#!/bin/bash

source common.sh

# even when symlinks exist for host keys, when their targets are missing the installation tries to
# generate the keys. the fix is to have the symlinks point to empty files during the installation.
enableDirtyHack() {
	mkdir -p /persist/apps/ssh-server

	touch /persist/apps/ssh-server/{ssh_host_ecdsa_key,ssh_host_ecdsa_key.pub,ssh_host_ed25519_key,ssh_host_ed25519_key.pub,ssh_host_rsa_key,ssh_host_rsa_key.pub}
}

cleanupDirtyHack() {
	rm -rf /persist/apps/ssh-server
}


enableDirtyHack

DEBIAN_FRONTEND=noninteractive apt install -y openssh-server

cleanupDirtyHack

#versioncommand: apt show openssh-server | grep Version: | cut -d' ' -f2
