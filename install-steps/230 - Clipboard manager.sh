#!/bin/bash

source common.sh


# a clipboard manager is required if you don't want clipboard to empty when the program exits where
# you copied the data from
curl -fsSL "https://github.com/xrelkd/clipcat/releases/download/v0.5.0/clipcat-v0.5.0-x86_64-unknown-linux-gnu.tar.gz" \
	| tar -xz -C /usr/bin -f - clipcat-menu clipcat-notify clipcatctl clipcatd
