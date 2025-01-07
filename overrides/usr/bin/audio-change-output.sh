#!/bin/bash -eu

sink="$(pactl list short sinks | cut -f 2 | rofi -dmenu -p "Change audio")";

pactl set-default-sink "$sink"
