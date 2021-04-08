#!/bin/bash

source common.sh


# it asks "Should non-superusers be able to capture packets?"
DEBIAN_FRONTEND=noninteractive apt install -y wireshark

#versioncommand: wireshark --version
