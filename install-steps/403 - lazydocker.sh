#!/bin/bash

source common.sh


# fantastic CLI UI for Docker
curl -fsSL \
	https://function61.com/app-dl/api/github.com/jesseduffield/lazydocker/latest_releases_asset/lazydocker_%2A_Linux_x86_64.tar.gz \
	| tar -C /usr/bin -xz lazydocker

#versioncommand: lazydocker --version | grep -oE 'Version: .+'
