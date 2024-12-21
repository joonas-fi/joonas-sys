#!/bin/bash

source common.sh


# Flatpak is a next-generation technology for building and distributing desktop applications on Linux
apt install -y flatpak

flatpak remote-add flathub https://flathub.org/repo/flathub.flatpakrepo

# all of this is stateful and needs to be stored in the persist partition.
# store this under "seed" name so the user can copy the content of this as the initial persist partition.
mv /var/lib/flatpak /var/lib/flatpak-seed
ln -s /sysroot/apps/flatpak /var/lib/flatpak

#versioncommand: flatpak --version | cut -d ' ' -f 2
