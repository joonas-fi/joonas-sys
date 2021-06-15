#!/bin/bash

source common.sh


# easy provisioning of development SSL certs

# dependency of mkcert
apt install -y libnss3-tools

downloadAndInstallSingleBinaryProgram \
	/usr/bin/mkcert \
	https://function61.com/app-dl/api/github.com/FiloSottile/mkcert/latest_releases_asset/mkcert-%2A-linux-amd64

#versioncommand: mkcert --version
