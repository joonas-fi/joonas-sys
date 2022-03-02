#!/bin/bash -eu

# many good things to add from: https://looks.wtf/

descriptionAndArt="$(cat /var/lib/rofimoji-data/ascii-art.txt | rofi -dmenu -p 'ASCII art')"

# "<description> <tab> <art>" => "<art>"
art="$(echo -n "$descriptionAndArt" | cut -d '	' -f 2-)"

# copy to clipboard
echo -n "$art" | xclip -selection clipboard

