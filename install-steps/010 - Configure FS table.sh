#!/bin/bash

source common.sh


mkdir -p /persist # mount point

# echo "/dev/sda1 / ext4  errors=remount-ro 0 1" > /etc/fstab
echo "LABEL=system_a  /         ext4  errors=remount-ro 0 1" > /etc/fstab
echo "LABEL=persist   /persist  ext4  errors=remount-ro 0 1" >> /etc/fstab
