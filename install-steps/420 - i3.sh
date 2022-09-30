#!/bin/bash

source common.sh


# tiling window manager (gaps fork for prettier visuals)


# status bar for i3. recommended would install dzen2 and i3-wm
# apt install --no-install-recommends -y i3status

# i3-gaps (a fork of i3 with gaps support) exists only as PPA
add-apt-repository -y ppa:regolith-linux/release

# a horrible hack to get packages for older release (b/c there are no packages for current release)
echo "deb https://ppa.launchpadcontent.net/regolith-linux/release/ubuntu impish main" > /etc/apt/sources.list.d/regolith-linux-ubuntu-release-jammy.list
apt update

# recommended would install dunst.
# suckless-tools = dmenu (used for workspace renaming)
# session means setting file so the i3 session shows up in greeter
apt install --no-install-recommends -y \
	i3-gaps-wm i3-gaps-session i3status i3lock suckless-tools

#versioncommand: i3 --version
