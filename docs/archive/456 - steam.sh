#!/bin/bash

source common.sh

# steam is fucking 32-bit https://www.quora.com/Why-is-Steam-32-bit
# so for steam, and only steam, we have to add support for 32-bit architecture (╯°□°)╯︵ ┻━┻
dpkg --add-architecture i386

# without this, steam fonts are garbled
# (EULA accept automation: https://askubuntu.com/a/25614)
echo ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true | debconf-set-selections
apt install -y ttf-mscorefonts-installer

# TODO
# apt install steam

# what incredible horseshit:
# - steam news - i.e. steam ads
# - fugly all-caps aesthetic
# - font scaling is f'd up
# - doesn't use desktop notification standard
# - skip or view updates don't do anything
