#!/bin/bash

source common.sh


# A really good mesh VPN

apt install -y wireguard

# installation force-creates this *directory* which overrides our symlink.
# remove it so it will be added back via our post-install touch-ups.
if [ ! -L /etc/wireguard ]; then # not a symlink?
	rm -rf /etc/wireguard
fi
