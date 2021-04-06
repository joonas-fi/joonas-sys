#!/bin/bash

source common.sh


# I don't know what this does and if it needs to be done, but it was mentioned in
# https://help.ubuntu.com/community/Installation/FromLinux#Debootstrap
dpkg-reconfigure -f noninteractive tzdata
