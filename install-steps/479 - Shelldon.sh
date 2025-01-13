#!/bin/bash

source common.sh


# Shelldon is your new Rust-powered buddy with GPT features!
# NOTE: remember to supply ENV OPENAI_API_KEY. create API key here: https://platform.openai.com/api-keys
curl -fsSL 'https://function61.com/app-dl/api/github.com/douglasmakey/shelldon/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64' \
	| tar -C /usr/bin/ --wildcards shelldon -xzf -

# doesn't have version command
#versioncommand: echo -n 'n/a'
