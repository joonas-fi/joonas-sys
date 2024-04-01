#!/bin/bash

source common.sh


# don't install recommends, because it installs "Ubuntu fan". bring in most of recommended manually
apt install --no-install-recommends -y docker.io git cgroupfs-mount pigz xz-utils

# TODO: install compose v2 from Docker to rid of Python dependency
apt install -y docker-compose

# add user to Docker group, so we don't need to "$ sudo ..." all docker commands
usermod -aG docker "$username"

# Docker is not enabled by default.
# (instead it is socket-activated, i.e. "$ docker ps" would start "always-up" services)
systemctl enable docker

# installation force-creates this *directory* which overrides our symlink.
# remove it so it will be added back via our post-install touch-ups.
if [ ! -L /etc/docker ]; then # not a symlink?
	rm -rf /etc/docker
fi

# install buildx (CLI plugin that extends the docker command with the full support of the features provided by BuildKit)

# /usr/lib/docker/cli-plugins/... is a location from which Docker autodiscovers CLI plugins:
#   https://github.com/docker/buildx?tab=readme-ov-file#manual-download

downloadAndInstallSingleBinaryProgram /usr/lib/docker/cli-plugins/docker-buildx \
	"https://function61.com/app-dl/api/github.com/docker/buildx/latest_releases_asset/buildx-%2A.linux-amd64"

#versioncommand: docker --version
