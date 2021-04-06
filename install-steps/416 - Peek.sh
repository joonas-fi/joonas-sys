#!/bin/bash

source common.sh


# a screen recorder (.gif, .mp4, ..)
apt install -y peek

# 'dbus-launch --exit-with-session' prefix: https://askubuntu.com/a/311988
su "$username" -c "dbus-launch --exit-with-session gsettings set com.uploadedlobster.peek persist-save-folder /tmp"
su "$username" -c "dbus-launch --exit-with-session gsettings set com.uploadedlobster.peek recording-framerate 24"
