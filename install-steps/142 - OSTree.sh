#!/bin/bash

source common.sh


# system for versioning updates of Linux-based operating systems

apt install -y ostree

#versioncommand: apt show ostree | grep Version: | cut -d' ' -f2
