#!/bin/bash

source common.sh


# bsdmainutils = hexdump
# usbutils = lsusb
# pciutils = lspci
# dnsutils = nslookup, dig
# imagemagick = convert
# psmisc = killall/pstree
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
	psmisc \
	exiftool \
	tree

#versioncommand: curl --version | head -1 | cut -d ' ' -f 2
