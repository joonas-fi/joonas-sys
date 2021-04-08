#!/bin/bash

source common.sh


# bsdmainutils = hexdump
# usbutils = lsusb
# pciutils = lspci
# dnsutils = nslookup, dig
# imagemagick = convert
apt install -y \
	htop \
	iotop \
	curl \
	wget \
	unzip \
	jq \
	pv \
	ncdu \
	imagemagick \
	vim \
	strace \
	pciutils \
	usbutils \
	bsdmainutils \
	dnsutils \
	nmap \
	exiftool \
	tree

#versioncommand: curl --version | head -1 | cut -d ' ' -f 2
