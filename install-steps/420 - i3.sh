#!/bin/bash

source common.sh


# tiling window manager

# we used to install gaps fork from PPA but it proved a nightmare to always search a new maintained
# fork for each Ubuntu release. luckily the gaps feature was finally accepted upstream.

# recommended would install dunst.
# suckless-tools = dmenu (used for workspace renaming)
# i3status = status bar for i3. recommended would install dzen2 and i3-wm
apt install --no-install-recommends -y \
	i3-wm i3status i3lock suckless-tools

#versioncommand: i3 --version
