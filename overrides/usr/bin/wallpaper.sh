#!/bin/bash -eu

cd /sto/id/p2xQgboD7cI

wallpaper="$(ls -1 | sort --random-sort | head -1)"

xwallpaper --zoom "$wallpaper"

notify-send "Wallpaper change" "$wallpaper"
