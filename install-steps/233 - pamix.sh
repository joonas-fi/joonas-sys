#!/bin/bash

source common.sh


# volume level meter for PulseAudio
apt install -y pamix

#versioncommand: apt show pamix | grep Version: | cut -d' ' -f2
