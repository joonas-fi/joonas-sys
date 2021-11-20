#!/bin/bash

source common.sh


# this was in the docs as a step, maybe debootstrap installs the version at time of major release,
# and upgrade gets us the packages that came after that?
apt update -q && apt upgrade -qy

#versioncommand: apt --version
