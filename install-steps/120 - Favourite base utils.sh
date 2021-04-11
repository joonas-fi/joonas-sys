#!/bin/bash

source common.sh


# bsdmainutils = hexdump
# usbutils = lsusb
# pciutils = lspci
# dnsutils = nslookup, dig
# imagemagick = convert
# psmisc = killall/pstree
# binutils = strings
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
	binutils \
	usbutils \
	bsdmainutils \
	dnsutils \
	nmap \
	psmisc \
	exiftool \
	tree

#versioncommand: curl --version | head -1 | cut -d ' ' -f 2
