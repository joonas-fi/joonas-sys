#!/bin/bash

source common.sh


# a clipboard manager is required if you don't want clipboard to empty when the program exits where
# you copied the data from
curl -fsSL \
	https://function61.com/app-dl/api/github.com/xrelkd/clipcat/latest_releases_asset/clipcat-%2A-x86_64-unknown-linux-gnu.tar.gz \
	| tar -xz -C /usr/bin -f - clipcat-menu clipcat-notify clipcatctl clipcatd

#versioncommand: clipcatd --version
