#!/bin/bash

source common.sh


# up is the Ultimate Plumber, a tool for writing Linux pipes in a terminal-based UI interactively, with instant live preview of command results.

# unconventional naming - the base name without any "-<platfrom>" suffix is for Linux
downloadAndInstallSingleBinaryProgram /usr/bin/up https://function61.com/app-dl/api/github.com/akavel/up/latest_releases_asset/up

#versioncommand: up --help 2>&1 | grep 'VERSION: ' | cut -d' ' -f2
