#!/bin/bash -eu

cd /sto/id/p2xQgboD7cI

wallpaper="$(ls -1 | sort --random-sort | head -1)"

mkdir -p /dev/shm/wallpaper-change

cp "$wallpaper" /dev/shm/wallpaper-change/

notify-send "Wallpaper change" "Resizing..."

docker run --rm --net=none -v /dev/shm/wallpaper-change:/workspace joonas/imagemagick convert "$wallpaper" wallpaper.png

xwallpaper --zoom /dev/shm/wallpaper-change/wallpaper.png

cp /dev/shm/wallpaper-change/wallpaper.png /sysroot/apps/SYSTEM/background.png

notify-send "Wallpaper change" "$wallpaper"

rm -rf /dev/shm/wallpaper-change
