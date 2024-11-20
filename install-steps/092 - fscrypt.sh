#!/bin/bash

source common.sh


# Tool for managing Linux filesystem encryption

apt install -y fscrypt

#versioncommand: apt show fscrypt | grep Version: | cut -d' ' -f2
