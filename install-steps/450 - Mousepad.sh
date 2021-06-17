#!/bin/bash

source common.sh


apt install -y mousepad

# use for .txt etc files (so they won't open in LibreOffice)
su "$username" -c "xdg-mime default mousepad.desktop text/plain"

#versioncommand: mousepad --version
