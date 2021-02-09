#!/bin/bash -eu

# TODO: use this
sysId="a"

volatilePersistPartition="/dev/shm/sys-${sysId}-persist-volatile"
# volatilePersistPartition="/mnt/system-provision/persist.img"

rootPartition="/dev/shm/joonas-os-ram-image"
# rootPartition="/dev/disk/by-label/system_${sysId}"

espPartition="/dev/disk/by-label/ESP"
# espPartition="/dev/sda1"
# espPartition="/mnt/system-provision/esp.img"


cowNameEsp="misc/vm-test-disks/esp.qcow2"
cowNameRoot="misc/vm-test-disks/root-ro.qcow2"

function startWithEmptyPersistPartition {
	rm -f "$volatilePersistPartition"

	# creates sparse file, i.e. does not allocate for empty sections
	truncate -s 4G "$volatilePersistPartition"

	mkfs.ext4 -L persist "$volatilePersistPartition" 2> /dev/null
}

function createReadonlyEspAndRootRoSystem {
	rm -f "$cowNameEsp" "$cowNameRoot"

	qemu-img create -f qcow2 -b "$espPartition" "$cowNameEsp"
	qemu-img create -f qcow2 -b "$rootPartition" "$cowNameRoot"
}

startWithEmptyPersistPartition

createReadonlyEspAndRootRoSystem

# the various OVMF_VARS files decide which system we'll boot automatically (UEFI vars remembering
# last selected boot option)

# RNG device supposedly speeds up Ubuntu boot
qemu-system-x86_64 \
	-machine type=q35,accel=kvm \
	-drive "file=${cowNameEsp}" \
	-drive "file=${cowNameRoot}" \
	-drive "format=raw,file=${volatilePersistPartition}" \
	-drive if=pflash,format=raw,unit=0,readonly,file=misc/uefi-files/OVMF_CODE-pure-efi.fd \
	-drive if=pflash,format=raw,unit=1,readonly,file="misc/uefi-files/OVMF_VARS-boot-system-${sysId}.fd" \
	-m 4G \
	-smp 4
