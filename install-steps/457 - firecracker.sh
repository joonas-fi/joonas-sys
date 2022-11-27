#!/bin/bash

source common.sh


# enables you to deploy workloads in lightweight virtual machines, called microVMs

curl -fsSL https://function61.com/app-dl/api/github.com/firecracker-microvm/firecracker/latest_releases_asset/firecracker-%2A-x86_64.tgz \
	| tar -C /usr/bin -xz --strip-components=1 --wildcards '*/firecracker-*-x86_64'

mv /usr/bin/firecracker-*-x86_64 /usr/bin/firecracker

#versioncommand: firecracker --version | head -1 | cut -d  ' ' -f2
