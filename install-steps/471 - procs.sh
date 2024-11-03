#!/bin/bash

source common.sh


# a replacement for ps written in Rust

installProcs() {
	rm -rf /tmp/procs-download
	mkdir /tmp/procs-download
	cd /tmp/procs-download
	curl --fail --location --show-error --no-progress-meter --output procs.zip 'https://function61.com/app-dl/api/github.com/dalance/procs/latest_releases_asset/__autodetect__.zip?os=linux&arch=amd64'
	unzip -j procs.zip procs -d /usr/bin/
	rm -rf /tmp/procs-download
}

installProcs

#versioncommand: procs --version | | grep -oP '(?<=procs ")[^ ]+'
