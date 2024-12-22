#!/bin/bash

# here's how i got it working
# ---------------------------
# 
# install gpu-manager (ubuntu-drivers-common)
# 
# run prime-switch
# 
# remove gdm3
# 
# source common.sh

# overridden by us. apt will freak out on package installation if this file already exists, so
# we remove it. the final steps will put it back.
rm /etc/gdm3/custom.conf

# (why "noninteractive"? read rant about gdm3 below)
#
# unless we specify slick-greeter, it's going to default to unity-greeter which pulls in
# parts of Unity also

# - gdm3 is a display manager (= GUI for logging in to your desktop)
# - rofi is an application launcher
# - xwallpaper might not be always required (once hautomo-client can set wallpapers without it)
# - ttf-ancient-fonts because emojis didn't render (https://www.omgubuntu.co.uk/2014/11/see-install-use-emoji-symbols-ubuntu-linux)
# - fonts-noto-color-emoji to get colored emojis for i3 workspace symbols
# - gvfs-backends = Samba working in thunar
# - gvfs-fuse = GVFS files to work in non-GIO (= POSIX) programs
# - xdg-utils = xdg-mime (for setting default programs for file types)
DEBIAN_FRONTEND=noninteractive apt install -y \
	sway \
	xdg-utils \
	gvfs-backends \
	gvfs-fuse \
	alsa \
	ttf-ancient-fonts \
	fonts-firacode \
	fonts-noto-color-emoji \
	fonts-hack \
	fonts-powerline

# before for some reason lightdm had to be installed in a separate "$ apt" call, otherwise gdm3
# would get pulled in. now after I removed something from here it gets pulled in anyway. JFC I
# tried to research which package pulls it, I couldn't figure it out and now we have conflict
# because we have 2 managers, and now we've to fix it. UGH!

# echo /usr/bin/lightdm > /etc/X11/default-display-manager
# 
# rm /etc/systemd/system/display-manager.service
# 
# ln -s /lib/systemd/system/lightdm.service /etc/systemd/system/display-manager.service

#versioncommand: lightdm --version
