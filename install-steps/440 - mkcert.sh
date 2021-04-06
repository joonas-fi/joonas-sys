#!/bin/bash

source common.sh


# easy provisioning of development SSL certs

# dependency of mkcert
apt install -y libnss3-tools

downloadAndInstallSingleBinaryProgram \
	/usr/bin/mkcert \
	"https://github.com/FiloSottile/mkcert/releases/download/v1.4.3/mkcert-v1.4.3-linux-amd64"
