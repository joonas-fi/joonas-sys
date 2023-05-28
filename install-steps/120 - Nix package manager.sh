#!/bin/bash

source common.sh

# "Nix is a tool that takes a unique approach to package management and system configuration.
# Learn how to make reproducible, declarative and reliable systems."


# Single-user installation
su "$username" -c "sh <(curl -L https://nixos.org/nix/install) --no-daemon"

# TODO: enable
#       versioncommand: nix --version | cut -d ' ' -f 3
