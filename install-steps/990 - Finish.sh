#!/bin/bash -eu

source common.sh


# re-do the same as "Copy overrides directory".
#
# this is because some software override our files. (Wireguard, Docker installation /etc/docker)
#
# as a hack, after some software have done this, after installation we remove their overrides and
# run this again so they get back.

rsync -v --exclude=.empty_dir -a "${repodir}/overrides/" /

# in case we put in place some fonts
fc-cache -f -v
