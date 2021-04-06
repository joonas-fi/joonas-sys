#!/bin/bash

source common.sh


# tiling window manager (gaps fork for prettier visuals)


# status bar for i3. recommended would install dzen2 and i3-wm
# apt install --no-install-recommends -y i3status

# i3-gaps (a fork of i3 with gaps support) exists in speed-ricer repo
add-apt-repository -y ppa:kgilmer/speed-ricer

# recommended would install dunst.
# suckless-tools = dmenu (used for workspace renaming)
# session means setting file so the i3 session shows up in greeter
apt install --no-install-recommends -y \
	i3-gaps-wm i3-gaps-session i3status suckless-tools