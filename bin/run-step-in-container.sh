#!/bin/bash -eu

# This is a small "proxy script" because there's quite a lot of ceremony in setting up
# mounts and chroot

# install-steps/<file>
installStepToRun="$1"

# bootstrap work prepares us being able to chroot to here
chrootLocation="/mnt/j-os-inmem-staging"

if [ ! -f "${chrootLocation}/etc/os-release" ]; then
	# --archive (same as -dR --preserve=all) needed to preserve symlinks etc (apt soils itself without this)
	cp --archive -r /debootstrap-cache/* "$chrootLocation"

	# debootstrap will end up with almost-empty sources list. copy list from
	# our builder Docker image
	cp /etc/apt/sources.list "${chrootLocation}/etc/apt/sources.list"

	echo "# Copied APT sources.list"
fi

# we have to do this after debootstrap, because it creates stuff in /dev (device node files),
# and for some reason we need these mounts because some installers access these
mount --bind /dev "${chrootLocation}/dev"
mount --bind /dev/pts "${chrootLocation}/dev/pts"
mount -t proc proc "${chrootLocation}/proc"
mount -t sysfs sys "${chrootLocation}/sys"

# access to our repo must be available inside the chroot
mkdir -p "${chrootLocation}/tmp/repo"
mount --bind /repo "${chrootLocation}/tmp/repo"

# chroot into our new root and continue installation there. it now has basic dependencies for starting
# to build a system by installing apt packages

# if we're interactive, just get a shell inside the chroot (aids in debugging problems & iterating faster)
if [ -t 0 ]; then
	echo "Dropping you to interactive chroot."
	chroot "$chrootLocation" bash
else
	# chroot sets workdir to (new) root, so we need "$ cd" to give steps the correct perspective
	chroot "$chrootLocation" sh -c "cd /tmp/repo/install-steps && ./'$installStepToRun'"
fi
