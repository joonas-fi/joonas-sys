#!/bin/bash

source common.sh


# Text editor


# instructions from https://www.sublimetext.com/docs/3/linux_repositories.html

echo "deb https://download.sublimetext.com/ apt/stable/" > /etc/apt/sources.list.d/sublime-text.list

wget -qO - https://download.sublimetext.com/sublimehq-pub.gpg | apt-key add -

apt update && apt install -y sublime-text

# alternate way with Snap:
# snap install --classic sublime-text
