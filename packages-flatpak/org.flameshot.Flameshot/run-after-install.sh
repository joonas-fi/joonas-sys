#!/bin/bash -eu

# without this the app doesn't see the host /tmp
flatpak override --user --filesystem=/tmp org.flameshot.Flameshot
