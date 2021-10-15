#!/bin/bash

source common.sh


apt install -y wpasupplicant

# interferes with per-interface supplicants (netplan related?)
# disabling is not enough, as this is maybe invoked from a dependency?
systemctl mask wpa_supplicant

#versioncommand: apt show wpasupplicant | grep Version: | cut -d' ' -f2
