#!/bin/bash

source common.sh


# A really good mesh VPN

apt install -y wireguard

# installation unfortunately replaces our symlink with an empty dir
if [ ! -L /etc/wireguard ]; then # not a symlink? => replace our symlink back
	rm -rf /etc/wireguard
	cp -a "${repodir}/overrides/etc/wireguard" /etc/wireguard
fi
