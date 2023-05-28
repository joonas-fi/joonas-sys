#!/bin/bash

source common.sh

installVirtiofsd() {
	mkdir /tmp/virtiofsd-download
	cd /tmp/virtiofsd-download
	# for some reason wget doesn't work (we get 403), but only inside a container
	curl -fsSL -o virtiofsd.zip "https://gitlab.com/virtio-fs/virtiofsd/-/jobs/artifacts/main/download?job=publish"
	unzip -j virtiofsd.zip target/x86_64-unknown-linux-musl/release/virtiofsd -d /usr/bin
	rm -rf /tmp/virtiofsd-download
}

# Virtual Machines
# also uses Samba (for VM shared folders), which we install earlier
# qemu-system-misc for RISC-V and some other exotic architectures.
# qemu-efi-aarch64 to have EFI firmware

apt install -y qemu-system-x86 qemu-system-arm qemu-system-misc \
	qemu-efi-aarch64

installVirtiofsd

#versioncommand: qemu-system-x86_64 --version
