#!/bin/bash

source common.sh


downloadAndInstallSingleBinaryProgram \
	/usr/bin/function22 \
	"https://function61.com/app-dl/api/github.com/function61/function22/latest_releases_asset/function22_linux-amd64"

# only listen on Tailscale interface
function22 install --allowed-users="$username" --interface=tailscale0 --validate-host-key=false

systemctl enable function22

#versioncommand: function22 --version | cut -d' ' -f3
