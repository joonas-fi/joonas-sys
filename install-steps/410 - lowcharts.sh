#!/bin/bash

source common.sh


# Tool to draw low-resolution graphs in terminal.

# only grab binary from inside of the archive (there is also LICENSE)

curl -fsSL \
	https://function61.com/app-dl/api/github.com/juan-leon/lowcharts/latest_releases_asset/lowcharts-%2A-x86_64-unknown-linux-gnu.tar.gz \
	| tar -C /usr/bin/ -xzf - lowcharts

#versioncommand: lowcharts --version | cut -d ' ' -f 2
