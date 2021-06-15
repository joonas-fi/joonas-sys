#!/bin/bash

source common.sh


# Centralized file storage
downloadAndInstallSingleBinaryProgram \
	/usr/bin/sto \
	https://function61.com/app-dl/api/github.com/function61/varasto/latest_releases_asset/sto_linux-amd64

#versioncommand: sto --version

# FUSE mount point needs to be owned by us
chown "$username:$username" /sto
