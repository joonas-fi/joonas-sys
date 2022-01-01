#!/bin/bash

source common.sh

# copy jsys command to the installed system
cp "${repodir}/rel/jsys_linux-amd64" /usr/bin/jsys

# install lowdiskspace-checker as systemd service and ..
su "$username" -c "jsys lowdiskspace-checker systemd-units"

# .. enable the timer
su "$username" -c "systemctl --user enable lowdiskspace-checker.timer"
