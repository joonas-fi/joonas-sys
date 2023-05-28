#!/bin/bash

source common.sh


# A hackable, minimal, fast TUI file explorer

curl -fsSL https://function61.com/app-dl/api/github.com/sayanarijit/xplr/latest_releases_asset/xplr-linux.tar.gz \
	| tar -C /usr/bin -xzf -

#versioncommand: xplr --version | cut -d ' ' -f2
