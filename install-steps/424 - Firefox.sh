#!/bin/bash

source common.sh


# libavcodec-extra b/c some website video players aren't working:
#   https://askubuntu.com/questions/1035661/playing-videos-in-firefox
apt install -y firefox libavcodec-extra xdg-utils

# https://askubuntu.com/questions/16621/how-to-set-the-default-browser-from-the-command-line
xdg-settings set default-web-browser firefox.desktop

#versioncommand: firefox --version
