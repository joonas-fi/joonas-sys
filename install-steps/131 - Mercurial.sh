#!/bin/bash

source common.sh


apt install -y mercurial

# to install hg-git we need pip first
# (credits https://stackoverflow.com/a/65125295)

(
	tempInstallDir="/tmp/pip-install"

	mkdir -p "$tempInstallDir" && cd "$tempInstallDir"
	curl -fsSL https://bootstrap.pypa.io/pip/2.7/get-pip.py --output get-pip.py
	python2 get-pip.py
	rm -rf "$tempInstallDir"
)

# https://pypi.org/project/hg-git/
# (clone from github didn't work without these additional modules)
pip install hg-git brotli ipaddress

#versioncommand: hg --version
