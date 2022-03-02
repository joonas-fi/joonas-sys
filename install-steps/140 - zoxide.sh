#!/bin/bash

source common.sh


# zoxide is a smarter cd command, inspired by z and autojump.
# It remembers which directories you use most frequently, so you can "jump" to them in just a few keystrokes.
# zoxide works on all major shells.

curl -fsSL \
	https://function61.com/app-dl/api/github.com/ajeetdsouza/zoxide/latest_releases_asset/zoxide-%2A-x86_64-unknown-linux-musl.tar.gz \
	| tar -C /usr/bin -xz zoxide

#versioncommand: zoxide --version | cut -d' ' -f2
