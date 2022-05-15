#!/bin/bash

source common.sh


# Alternative for sudo

apt install -y doas

#versioncommand: apt show doas | grep Version: | cut -d' ' -f2
