#!/bin/bash

source common.sh


# Alternative to Systemd. used with Unikie project for a while.
apt install -y supervisor

#versioncommand: apt show supervisor | grep Version: | cut -d' ' -f2
