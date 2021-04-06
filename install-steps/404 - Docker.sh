#!/bin/bash

source common.sh


# don't install recommends, because it installs "Ubuntu fan". bring in most of recommended manually
apt install --no-install-recommends -y docker.io git cgroupfs-mount pigz xz-utils

apt install -y docker-compose

# add user to Docker group, so we don't need to "$ sudo ..." all docker commands
usermod -aG docker "$username"

# Docker is not enabled by default.
# (instead it is socket-activated, i.e. "$ docker ps" would start "always-up" services)
systemctl enable docker
