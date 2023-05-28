#!/bin/bash

source common.sh


# A post-modern modal text editor.

curl -fsSL https://function61.com/app-dl/api/github.com/helix-editor/helix/latest_releases_asset/helix-%2A-x86_64-linux.tar.xz \
	| tar --strip-components=1 -C /usr/bin -xJf - --wildcards '*/hx'

#versioncommand: hx --version | cut -d ' ' -f2

