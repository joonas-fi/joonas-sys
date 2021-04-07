#!/bin/bash

source common.sh


# (why "noninteractive"? read rant about gdm3 below)
#
# unless we specify slick-greeter, it's going to default to unity-greeter which pulls in
# parts of Unity also

# - lightdm is a display manager (= GUI for logging in to your desktop)
# - rofi is an application launcher
# - xwallpaper might not be always required (once hautomo-client can set wallpapers without it)
# - ttf-ancient-fonts because emojis didn't render (https://www.omgubuntu.co.uk/2014/11/see-install-use-emoji-symbols-ubuntu-linux)
# - fonts-noto-color-emoji to get colored emojis for i3 workspace symbols
DEBIAN_FRONTEND=noninteractive apt install -y \
	xfce4 \
	xfce4-screensaver \
	xfce4-notifyd \
	lightdm \
	slick-greeter \
	alsa \
	compton \
	xwallpaper \
	ttf-ancient-fonts \
	fonts-firacode \
	fonts-noto-color-emoji \
	fonts-hack \
	fonts-powerline

# before for some reason lightdm had to be installed in a separate "$ apt" call, otherwise gdm3
# would get pulled in. now after I removed something from here it gets pulled in anyway. JFC I
# tried to research which package pulls it, I couldn't figure it out and now we have conflict
# because we have 2 managers, and now we've to fix it. UGH!

echo /usr/bin/lightdm > /etc/X11/default-display-manager

rm /etc/systemd/system/display-manager.service

ln -s /lib/systemd/system/lightdm.service /etc/systemd/system/display-manager.service
