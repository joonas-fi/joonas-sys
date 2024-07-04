#!/bin/bash

source common.sh


# Interactive JSON filter using jq

curl --fail --show-error --no-progress-meter --location 'https://function61.com/app-dl/api/github.com/ynqa/jnv/latest_releases_asset/__autodetect__linux-gnu.tar.xz?os=linux&arch=amd64' \
	| tar -C /usr/bin --wildcards 'jnv-*/jnv' --strip-components=1 -xJf -

#versioncommand: jnv --version | cut -d ' ' -f2
