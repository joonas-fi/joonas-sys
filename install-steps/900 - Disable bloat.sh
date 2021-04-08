#!/bin/bash

source common.sh


# I don't care about manual regeneration
systemctl disable man-db.timer

# log rotation unnecessary for short-lived systems
systemctl disable logrotate.timer

# MOTD needn't be updated in a short-lived system
systemctl disable motd-news.timer

# APT package metadata needn't updated in a short-lived system.
# https://askubuntu.com/questions/1038923/do-i-really-need-apt-daily-service-and-apt-daily-upgrade-service
systemctl disable apt-daily.timer
systemctl disable apt-daily-upgrade.timer

# https://help.ubuntu.com/community/AutomaticSecurityUpdates , https://wiki.debian.org/UnattendedUpgrades
apt remove -y unattended-upgrades
