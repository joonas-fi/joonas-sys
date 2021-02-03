#!/bin/bash -eu

# system-a or system-b
sysId="a"

mountpoint="/mnt/system-${sysId}-maint"

# espMountPoint="/mnt/vm-efi"
espMountPoint="/boot/efi"

ramDiskImagePath="/dev/shm/joonas-os-ram-image"

function mountRealDisk() {
	mkdir -p "$mountpoint"

	# mount -o loop "$imagePath" "$mountpoint"
	mount "/dev/disk/by-label/system_${sysId}" "$mountpoint"
}

function mountRamDisk() {
	mkdir -p "$mountpoint"

	if [ -e "${mountpoint}/boot" ]; then
		unmountDisk
	fi

	# it's better to ensure we start from scratch. truncate + mkfs effectively reset the filesystem, sure,
	# but with sparse file we'll probably waste space if we don't do "hard reset"
	rm -f "$ramDiskImagePath"

	# creates sparse file, i.e. does not allocate for empty sections
	truncate -s 20G "$ramDiskImagePath"

	mkfs.ext4 -L "system_${sysId}" "$ramDiskImagePath"

	mount -o loop "$ramDiskImagePath" "$mountpoint"
}

function copySystreeToDisk() {
	# TODO: why is -ah different than "--acls --human-readable"?
	rsync -ah --delete --info=progress2 /mnt/j-os-inmem-staging/ "${mountpoint}/"
}

function stampSysIdIntoPersistPartition() {
	echo -n "${sysId}" > "${mountpoint}/etc/sys-id"
}

function copyKernelAndInitrd() {
	if [ ! -e "${espMountPoint}/EFI" ]; then
		echo "ESP not mounted"
		return 1
	fi

	# intentionally failing for safety if parent dir does not exist

	local espDir="${espMountPoint}/EFI/system${sysId}"

	# our UEFI setup: we could add additional arguments to vmlinuz's boot entry, but those go to
	# motherboards's NVRAM and thus is not portable if you take the disk to another computer. we write
	# a simple UEFI shell script (boot.nsh) instead which contains the kernel command line and use
	# that as the boot entry, so the bootable parameters are self-contained on the ESP

	cp "${mountpoint}/boot/vmlinuz" "${espDir}/vmlinuz"
	cp "${mountpoint}/boot/initrd.img" "${espDir}/initrd.img"

	# unfortunately we have to refer to vmlinuz with full path, because there's concept of current
	# workdir and it's not set to the script when invoked..
	echo "\\EFI\\system${sysId}\\vmlinuz root=LABEL=system_${sysId} initrd=\\EFI\\system${sysId}\\initrd.img" > "${espDir}/boot.nsh"
}

function unmountDisk() {
	umount "$mountpoint"
	# rm -r "$mountpoint"
}

function unmountIfDangling() {
	if [ -e "$mountpoint" ]; then
		echo "WARN: mountpoing left dangling from previous job. cleaning up"
		unmountDisk
	fi
}

function flashToActualPartition {
	# unmountIfDangling

	mountRealDisk

	copySystreeToDisk

	stampSysIdIntoPersistPartition

	copyKernelAndInitrd

	unmountDisk
}

function flashToInRamDisk {
	mountRamDisk

	copySystreeToDisk

	stampSysIdIntoPersistPartition

	copyKernelAndInitrd

	unmountDisk
}

flashToInRamDisk
