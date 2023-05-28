#!/bin/bash

source common.sh


# Ubuntu seems to add pipewire by default (still installing it here for explicitness), but
# pipewire-pulse needed to take over PulseAudio
DEBIAN_FRONTEND=noninteractive apt install -y pipewire pipewire-pulse wireplumber

#versioncommand: apt show pipewire | grep Version: | cut -d' ' -f2
