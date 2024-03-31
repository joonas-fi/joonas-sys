#!/bin/bash

source common.sh

# So that we can connect to a remote CUPS server (in a container) where the prints are actually handled.

apt install -y cups-client

#versioncommand: apt show cups-client | grep Version: | cut -d' ' -f2
