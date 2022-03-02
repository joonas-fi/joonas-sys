#!/bin/bash -eu

emojiAndDescription="$(cat /var/lib/rofimoji-data/* | rofi -dmenu -p 'Emoji')"

# "<emoji> <description>" => "<emoji>"
emoji="$(echo -n "$emojiAndDescription" | cut -d ' ' -f 1)"

# copy to clipboard
echo -n "$emoji" | xclip -selection clipboard

