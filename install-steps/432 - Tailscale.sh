#!/bin/bash

source common.sh

# Tailscale makes secure networking easy

# to get "$VERSION_CODENAME" i.e. Ubuntu release (focal/jammy/oracular...)
. /etc/os-release

curl -fsSL -o /usr/share/keyrings/tailscale-archive-keyring.gpg https://pkgs.tailscale.com/stable/ubuntu/$VERSION_CODENAME.noarmor.gpg
curl -fsSL -o /etc/apt/sources.list.d/tailscale.list https://pkgs.tailscale.com/stable/ubuntu/$VERSION_CODENAME.tailscale-keyring.list

apt update
apt install -y tailscale

#versioncommand: tailscale --version | head -1
