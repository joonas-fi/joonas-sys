#!/bin/bash

source common.sh


# bsdmainutils = hexdump
# usbutils = lsusb
# binutils = strings
apt install -y \
	curl \
	wget \
	unzip \
	jq \
	pv \
	vim \
	strace \
	binutils \
	usbutils \
	bsdmainutils \
	tree

#versioncommand: curl --version | head -1 | cut -d ' ' -f 2
