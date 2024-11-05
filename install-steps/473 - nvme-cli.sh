#!/bin/bash

source common.sh


# NVMe management command line interface.

apt install -y nvme-cli

#versioncommand: apt show nvme-cli | grep Version: | cut -d' ' -f2
