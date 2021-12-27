#!/bin/bash

source common.sh


# a general-purpose command-line fuzzy finder

curl -fsSL \
	https://function61.com/app-dl/api/github.com/junegunn/fzf/latest_releases_asset/fzf-%2A-linux_amd64.tar.gz \
	| tar -C /usr/bin/ -xzf -

#versioncommand: fzf --version
