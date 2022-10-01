#!/bin/bash -eu

add() {
	local package="$1"

	flatpak install --assumeyes flathub "$package"

	# package might have some actions to be ran after install (such as writing settings programmatically)
	if [ -f "/etc/packages-flatpak/$package/run-after-install.sh" ]; then
		"/etc/packages-flatpak/$package/run-after-install.sh"
	fi
}

# dir names in /etc/packages-flatpak/ indicate which packages we should install
for package_with_path in /etc/packages-flatpak/*
do
	# "packages-flatpak/com.brave.Browser" -> "com.brave.Browser"
	package="$(basename $package_with_path)"

	add "$package"
done
