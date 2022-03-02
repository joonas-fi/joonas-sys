#!/bin/bash

source common.sh


# a VNC server for real X displays

apt install -y x11vnc

#versioncommand: apt show x11vnc | grep Version: | cut -d' ' -f2
