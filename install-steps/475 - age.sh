#!/bin/bash

source common.sh


# A simple, modern and secure encryption tool

curl -fsSL 'https://function61.com/app-dl/api/github.com/FiloSottile/age/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64' \
	| tar -C /usr/bin/ --wildcards 'age/age' 'age/age-keygen' --strip-components=1 -xzf -

#versioncommand: age --version
