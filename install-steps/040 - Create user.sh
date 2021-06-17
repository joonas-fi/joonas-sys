#!/bin/bash

source common.sh


# each user should have a desktop directory (this is copied by useradd)
mkdir -p /etc/skel/Desktop

# order of "GECOS" (AKA comment) field: https://askubuntu.com/a/94067
#  how to generate $USER_PASSWORD_HASH: https://askubuntu.com/a/667842
useradd \
	--create-home \
	--password "$USER_PASSWORD_HASH" \
	--shell /bin/bash \
	--comment "Joonas" \
	"$username"

# allow use of (modems?) Zigbee USB stick
usermod -a -G dialout "$username"

# allow reading journals
usermod -a -G systemd-journal "$username"

# add as sudoer
gpasswd --add "$username" sudo

# do not require re-auth when invoking "$ sudo ..."
echo "$username ALL=(ALL) NOPASSWD: ALL" > "/etc/sudoers.d/$username" \
	&& chmod 440 "/etc/sudoers.d/$username"
