#!/bin/bash

source common.sh


# Distribute Artifacts Across OCI Registries With Ease

curl -fsSL 'https://function61.com/app-dl/api/github.com/oras-project/oras/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64' \
	| tar -C /usr/bin/ --wildcards 'oras' -xzf -

#versioncommand: oras version | grep -oP 'Version:\s+\K[\d.]+'
