#!/bin/bash

source common.sh


apt install -y rsync

# .empty_dir b/c Git can't track empty directories (without placing dummy file in it), but sometimes
# an actual empty directory is important (like mount points), so this exclude gets rsync to make empty dirs
rsync --exclude=.empty_dir -a "${repodir}/overrides/" /

# make /etc/packages-flatpak based on our repo
rsync -a "${repodir}/packages-flatpak" /etc/

# any file we wrote under user's home dir, we wrote as root
chown -R "$username:$username" "/home/$username"

# for some reason root ends up with other:write. was it because rsync?
chmod o-w /
