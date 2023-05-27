#!/bin/bash -eu

emojiAndDescription="$(cat /var/lib/rofimoji-data/* | rofi -dmenu -p 'Emoji')"

# unfortunately we have in this directory two different file "formats":
# 1. emojis with emoji first then description
# 2. ascii artwork with description first, then artwork

# dirty hack: all emoji descriptions contain "<small>" substring
if [[ $emojiAndDescription == *"<small>"* ]]; then
	# "<emoji> <description>" => "<emoji>"
	emoji="$(echo -n "$emojiAndDescription" | cut -d ' ' -f 1)"
else
	# "<description> <tab> <ascii art>" => "<ascii art>"
	emoji="$(echo -n "$emojiAndDescription" | cut -d '	' -f 2)"
fi

# copy to clipboard
echo -n "$emoji" | xclip -selection clipboard
