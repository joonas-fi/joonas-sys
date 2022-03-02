#!/bin/bash

source common.sh


# Flatpak is a next-generation technology for building and distributing desktop applications on Linux
apt install -y flatpak

flatpak remote-add flathub https://flathub.org/repo/flathub.flatpakrepo

#versioncommand: flatpak --version | cut -d ' ' -f 2
