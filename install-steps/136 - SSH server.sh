#!/bin/bash

source common.sh


downloadAndInstallSingleBinaryProgram \
	/usr/bin/function22 \
	"https://function61.com/app-dl/api/github.com/function61/function22/latest_releases_asset/function22_linux-amd64"

# this is an existing symlink for us
mv /etc/ssh/ssh_host_ed25519_key{,.temp}

# a hack because install requires the host key to be present at service installation time
#
# TODO: add --validate-host-key=false (once lands in a release) to overcome hack
function22 host-key-generate

# only listen on Tailscale interface
function22 install --allowed-users="$username" --interface=tailscale0

systemctl enable function22

# replace dummy-generated-file with our symlink
mv /etc/ssh/ssh_host_ed25519_key{.temp,}

#versioncommand: function22 --version | cut -d' ' -f3
