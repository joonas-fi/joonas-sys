#!/bin/bash

source common.sh


# A tool for glamorous shell scripts 🎀

curl -fsSL 'https://function61.com/app-dl/api/github.com/charmbracelet/gum/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64' \
	| tar -C /usr/bin --wildcards '*/gum' --strip-components=1 -xzf -

# update the "man page DB"
tldr --update

#versioncommand: gum --version | cut -d ' ' -f3
