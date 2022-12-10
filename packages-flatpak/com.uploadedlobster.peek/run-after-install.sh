#!/bin/bash -eu

echo 'gsettings set com.uploadedlobster.peek recording-framerate 24' | flatpak run '--command=sh' com.uploadedlobster.peek
echo 'gsettings set com.uploadedlobster.peek recording-output-format mp4' | flatpak run '--command=sh' com.uploadedlobster.peek

# TODO: need to give access to Flatpak to /tmp
# echo 'gsettings set com.uploadedlobster.peek persist-save-folder /tmp' | flatpak run '--command=sh' com.uploadedlobster.peek
