
# since all steps scripts source this file, they all "inherit" this mode which makes execution
# stop on errors
set -eu

# This installation process outline was shaped by:
# 	https://help.ubuntu.com/community/Installation/FromLinux#Debootstrap

# 0xx -> Low-level system stuff
# 1xx -> Base utils
# 2xx -> GUI base
# 4xx -> User-level applications
# 9xx -> Finishing touches

username="joonas"

# directory this repo is mounted inside in the build container
repodir="/tmp/repo"

source "${repodir}/secrets.env"

# shared utilities


function downloadAndInstallSingleBinaryProgram { # Go 😍
	local destination="$1"
	local downloadUrl="$2"

	curl --location --fail --silent --output "$destination" "$downloadUrl"

	chmod +x "$destination"
}
