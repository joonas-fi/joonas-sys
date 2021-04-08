#!/bin/bash

source common.sh


# fantastic CLI UI for Git
curl -fsSL "https://github.com/jesseduffield/lazygit/releases/download/v0.23.7/lazygit_0.23.7_Linux_x86_64.tar.gz" \
	| tar -C /usr/bin -xz lazygit

#versioncommand: lazygit --version | grep -oE 'version=[^ ,]+'
