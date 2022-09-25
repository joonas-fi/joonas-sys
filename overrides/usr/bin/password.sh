#!/bin/bash -eu

# check if not called from terminal => open in terminal window
if [ ! -t 0 ]; then
	i3-sensible-terminal --command "$0"
	exit 0
fi

cd /home/joonas/Desktop/Salasanoja

file=$(find . | fzf)

cat "$file"
read dummy # enter to exit reading the file

