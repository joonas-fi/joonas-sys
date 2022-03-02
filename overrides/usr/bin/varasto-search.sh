#!/bin/bash -eu

# refresh with:
#    sto debug build-collection-index > ~/.config/varasto/collection-index.txt

idAndPath="$(cat ~/.config/varasto/collection-index.txt | rofi -dmenu -i -p 'Varasto')"

# "<id> <path>" => "<id>"
id="$(echo -n "$idAndPath" | cut -d ' ' -f 1)"

sensible-browser "https://varasto.home.fn61.net/coll/${id}"

