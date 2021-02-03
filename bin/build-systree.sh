#!/bin/bash -eu

treeLocation="/mnt/j-os-inmem-staging"

# we can't use /dev/shm because it usually has "nodev" flag and Debian's booststrap process requires
# creating device files
function mountInRamFilesystem() {
	mkdir "$treeLocation"

	mount -t tmpfs -o size=16g tmpfs "$treeLocation"
}

if [ -d "$treeLocation" ]; then
	if [ "${1:-}" == "--rm" ]; then
		# quiet = "suppress 'not mounted' error messages". it still sets error status though..
		umount --quiet "$treeLocation" || true
		rm -rf "$treeLocation"

		mountInRamFilesystem
	elif [ "${1:-}" == "--keep" ]; then
		echo "Keeping current tree"
	else
		echo "Current systree exists. Run this script with '--rm' to umount-and-remove it (or use '--keep')"
		exit 1
	fi
else
	mountInRamFilesystem
fi

docker build -t j-os-builder .

# for "slave", see https://docs.docker.com/storage/bind-mounts/#configure-bind-propagation
docker run --rm -v "${treeLocation}:${treeLocation}:slave" --privileged j-os-builder
