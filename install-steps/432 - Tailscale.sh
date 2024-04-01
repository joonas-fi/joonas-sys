#!/bin/bash

source common.sh

# Tailscale makes secure networking easy

curl -fsSL -o /usr/share/keyrings/tailscale-archive-keyring.gpg https://pkgs.tailscale.com/stable/ubuntu/focal.noarmor.gpg
curl -fsSL -o /etc/apt/sources.list.d/tailscale.list https://pkgs.tailscale.com/stable/ubuntu/focal.tailscale-keyring.list

apt update
apt install -y tailscale

#versioncommand: tailscale --version | head -1
