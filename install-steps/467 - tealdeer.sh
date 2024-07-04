#!/bin/bash

source common.sh


# A very fast implementation of tldr in Rust.

downloadAndInstallSingleBinaryProgram /usr/bin/tldr "https://function61.com/app-dl/api/github.com/dbrgn/tealdeer/latest_releases_asset/__autodetect__-musl?os=linux&arch=amd64"

# update the "man page DB"
su "$username" -c "tldr --update"


#versioncommand: tldr --version | cut -d ' ' -f2
