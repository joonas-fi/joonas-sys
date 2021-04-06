#!/bin/bash

source common.sh


# CLI filesystem navigator

curl -fsSL https://github.com/gokcehan/lf/releases/download/r19/lf-linux-amd64.tar.gz \
	| tar -C /usr/bin/ -xzf -
