#!/bin/bash

source common.sh


# libavcodec-extra b/c some website video players aren't working:
#   https://askubuntu.com/questions/1035661/playing-videos-in-firefox
apt install -y firefox libavcodec-extra

xdg-settings set default-web-browser firefox.desktop
