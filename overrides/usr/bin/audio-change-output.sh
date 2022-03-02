#!/bin/bash -eu

sink="$(pactl list short sinks | cut -f 2 | rofi -dmenu -p "Change audio")";

pacmd set-default-sink "$sink"
