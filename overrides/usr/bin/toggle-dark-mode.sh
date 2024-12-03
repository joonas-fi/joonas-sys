#!/bin/bash -eu

current_color="$(gsettings get org.gnome.desktop.interface color-scheme)"

if [[ "$current_color" == *"dark"* ]]; then
	# go light
	gsettings set org.gnome.desktop.interface color-scheme 'default'
else
	# go dark
	gsettings set org.gnome.desktop.interface color-scheme 'prefer-dark'
fi

