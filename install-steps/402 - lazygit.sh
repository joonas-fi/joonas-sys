#!/bin/bash

source common.sh


# fantastic CLI UI for Git
curl -fsSL \
	https://function61.com/app-dl/api/github.com/jesseduffield/lazygit/latest_releases_asset/lazygit_%2A_Linux_x86_64.tar.gz \
	| tar -C /usr/bin -xz lazygit

#versioncommand: lazygit --version | grep -oE 'version=[^ ,]+'
