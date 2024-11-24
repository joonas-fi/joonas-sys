#!/bin/bash

source common.sh


installRclone() {
	mkdir /tmp/rclone-install
	cd /tmp/rclone-install
	curl -fsSL -o rclone.zip "https://function61.com/app-dl/api/github.com/rclone/rclone/latest_releases_asset/__autodetect__.zip?os=linux&arch=amd64"
	unzip -j rclone.zip 'rclone-*/rclone' -d /usr/bin
	rm -rf /tmp/rclone-install
}

# rsync for cloud storage

installRclone

#versioncommand: rclone --version | cut -d' ' -f2 | head -1
