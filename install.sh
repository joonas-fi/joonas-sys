#!/bin/bash -eu

username="joonas"

# for how to generate https://askubuntu.com/a/667842
passwordHash='TODO'

timezone="Europe/Helsinki"

repodir="/tmp/repo"


# This installation process was much shaped by:
# 	https://help.ubuntu.com/community/Installation/FromLinux#Debootstrap


# ----- utilities -----

function downloadAndInstallSingleBinaryProgram { # Go 😍
	local destination="$1"
	local downloadUrl="$2"

	curl --location --fail --silent --output "$destination" "$downloadUrl"

	chmod +x "$destination"
}

# ----- /utilities -----

function configureFilesystemTable {
	mkdir -p /persist # mount point

	# echo "/dev/sda1 / ext4  errors=remount-ro 0 1" > /etc/fstab
	echo "LABEL=system_a  /         ext4  errors=remount-ro 0 1" > /etc/fstab
	echo "LABEL=persist   /persist  ext4  errors=remount-ro 0 1" >> /etc/fstab
}

function setupNetwork {
	echo "work" > /etc/hostname

	# https://wiki.archlinux.org/index.php/systemd-networkd

	# for my actual machine
	echo -e "[Match]\nName=enp3s0\n\n[Network]\nDHCP=yes\n" > /etc/systemd/network/20-enp3s0.network

	# for testing in VM
	echo -e "[Match]\nName=enp0s2\n\n[Network]\nDHCP=yes\n" > /etc/systemd/network/21-enp0s2.network

	# for some reason the network configuration daemon is not up by default
	systemctl enable systemd-networkd

	# mkdir -p /etc/network

	# echo -e "auto eth0\niface eth0 inet dhcp" > /etc/network/interfaces
	# echo -e "auto enp2s0\niface enp0s2 inet dhcp" > /etc/network/interfaces
}

function configureTimezoneAndLocale {
	rm /etc/localtime
	ln -s "/usr/share/zoneinfo/$timezone" /etc/localtime

	echo "$timezone" > /etc/timezone

	# I don't know what this does and if it needs to be done, but it was mentioned in
	# https://help.ubuntu.com/community/Installation/FromLinux#Debootstrap
	dpkg-reconfigure -f noninteractive tzdata
}

# this was in the docs as a step, maybe debootstrap installs the version at time of major release,
# and upgrade gets us the packages that came after that?
function aptUpdateAndUpgrade {
	apt update && apt upgrade -y
}

function aptEnableAddAptRepository {
	# apparently we need to download 42 MB of packages (including GPG, pinentry and gstreamer!!!)
	# to be able to get stuff from PPA repositories
	apt install -y software-properties-common
}

# don't know why, but I guess it's needed
function installLanguagePack {
	apt install -y language-pack-en-base
}

function installLvm {
	apt install -y lvm2
}

function installKernelAndBootloader {
	# generic = not lowlatency or some other specific 
	# noninteractive because GRUB complains about bootsector
	# DEBIAN_FRONTEND=noninteractive apt install -y grub-pc linux-image-generic

	# hmm maybe we don't need GRUB
	DEBIAN_FRONTEND=noninteractive apt install -y linux-image-generic

	# have mount point ready for ESP
	mkdir -p /boot/efi
}

function setupRootOverlay {
	# instruct pre-boot environment to have overlay kernel module loaded
	echo "overlay" >> /etc/initramfs-tools/modules

	# update needed after we modified the contents (later steps probably do this, but it'd be dirty to rely on it)
	update-initramfs -u -k all
}

function installGpuDriver {
	# somewhere along the process AMD GPU drivers seem to get automatically pulled in..
	# nothing explicit implemented.
	return 0
}

function createUser {
	# each user should have a desktop directory (this is copied by useradd)
	mkdir -p /etc/skel/Desktop

	# order of "GECOS" (AKA comment) field: https://askubuntu.com/a/94067
	useradd \
		--create-home \
		--password "$passwordHash" \
		--shell /bin/bash \
		--comment "Joonas" \
		"$username"

	# add as sudoer
	gpasswd --add "$username" sudo

	# do not require re-auth when invoking "$ sudo ..."
	echo "$username ALL=(ALL) NOPASSWD: ALL" > "/etc/sudoers.d/$username" \
		&& chmod 440 "/etc/sudoers.d/$username"
}

# overlays the file/config override hierarchy on top of the root filesystem.
# this has to be done relatively early, because some steps depend on these (like initramfs generation)
function overlayOverrides() {
	apt install -y rsync

	rsync -a "${repodir}/overrides/" /
}

# snap/snapcraft is Docker-like but mainly focused for GUI apps
function installSnapd {
	apt install -y snapd
}

function installFavouriteBaseUtils {
	# bsdmainutils = hexdump
	# usbutils = lsusb
	# pciutils = lspci
	apt install -y \
		htop \
		curl \
		wget \
		unzip \
		jq \
		pv \
		ncdu \
		vim \
		strace \
		pciutils \
		usbutils \
		bsdmainutils \
		exiftool \
		tree
}

function installGit {
	apt install -y git
}

function installMercurial {
	apt install -y mercurial

	# to install hg-git we need pip first
	# (credits https://stackoverflow.com/a/65125295)

	(
		local tempInstallDir="/tmp/pip-install"

		mkdir -p "$tempInstallDir" && cd "$tempInstallDir"
		curl https://bootstrap.pypa.io/2.7/get-pip.py --output get-pip.py
		python2 get-pip.py
		rm -rf "$tempInstallDir"
	)

	# https://pypi.org/project/hg-git/
	# (clone from github didn't work without these additional modules)
	pip install hg-git brotli ipaddress
}

function installLazygit {
	curl -fsSL "https://github.com/jesseduffield/lazygit/releases/download/v0.23.7/lazygit_0.23.7_Linux_x86_64.tar.gz" \
		| tar -C /usr/bin -xz lazygit
}

function installDocker {
	# persist Docker data outside of our special copy-on-write root tree
	# mkdir -p /persist/docker-data

	ln -s /persist/docker-data /var/lib/docker

	apt install -y docker.io docker-compose

	# add user to Docker group, so we don't need to "$ sudo ..."
	# all docker commands
	usermod -aG docker "$username"
}

function installQemu {
	apt install -y qemu-system-x86
}

function installLf {
	curl -fsSL https://github.com/gokcehan/lf/releases/download/r19/lf-linux-amd64.tar.gz \
		| tar -C /usr/bin/ -xzf -
}

function installGraphicalEnvironment {
	# - dunst implements desktop notifications
	# - rofi is an application launcher
	# - xwallpaper might not be always required (once hautomo-client can set wallpapers without it)
	# - autorandr so each time a monitor disconnects the monitor configuration doesn't get lost
	# - xfce4-clipman is a clipboard manager, to be able to copy from programs we close before pasting
	# - ttf-ancient-fonts because emojis didn't render (https://www.omgubuntu.co.uk/2014/11/see-install-use-emoji-symbols-ubuntu-linux)
	apt install -y \
		xfce4 \
		xfce4-clipman \
		xfce4-screensaver \
		xfce4-terminal \
		xfce4-screenshooter \
		alsa \
		dunst \
		compton \
		xwallpaper \
		rofi \
		mousepad \
		ttf-ancient-fonts \
		fonts-firacode \
		fonts-hack \
		fonts-powerline
}

function installPeek {
	apt install -y peek

	su "$username" -c "dconf write /com/uploadedlobster/peek/persist-save-folder \"'/tmp'\""

	su "$username" -c "dconf write /com/uploadedlobster/peek/recording-framerate 24"
}

function installAutorandr {
	apt install -y autorandr
}

# graphical session manager = root user displaying GUI for user to login
function installGraphicalSessionManager {
	# for some reason if we install this alongside with xfce4 et al., (it yells about which to use,
	# gdm3 vs lightdm, even though gdm3 isn't installed by default if we don't ask for lightdm)
	apt install -y lightdm
}

function installI3Gaps {
	# status bar for i3
	apt install -y i3status

	# i3-gaps (a fork of i3 with gaps support) exists in speed-ricer repo
	add-apt-repository -y ppa:kgilmer/speed-ricer

	apt install -y i3-gaps
}

function installSublimeText {
	# instructions from https://www.sublimetext.com/docs/3/linux_repositories.html

	echo "deb https://download.sublimetext.com/ apt/stable/" > /etc/apt/sources.list.d/sublime-text.list

	wget -qO - https://download.sublimetext.com/sublimehq-pub.gpg | apt-key add -

	apt update && apt install -y sublime-text

	# alternate way with Snap:
	# snap install --classic sublime-text
}

function installFirefox {
	# libavcodec-extra b/c some website video players aren't working:
	#   https://askubuntu.com/questions/1035661/playing-videos-in-firefox
	apt install -y firefox libavcodec-extra
}

function installLibreoffice {
	apt install -y libreoffice-calc libreoffice-writer
}

function installCroc {
	(
		local tempInstallDir="/tmp/croc-install"

		mkdir "$tempInstallDir" && cd "$tempInstallDir"
		curl -fsSL -o croc.deb https://github.com/schollz/croc/releases/download/v8.6.7/croc_8.6.7_Linux-64bit.deb
		dpkg -i croc.deb 
		rm -rf "$tempInstallDir"
	)
}

function installVarasto {
	downloadAndInstallSingleBinaryProgram /usr/bin/sto "https://github.com/function61/varasto/releases/download/20200626_1423_4cd3ecf8/sto_linux-amd64"
}

function installHautomoClient {
	# downloadAndInstallSingleBinaryProgram /usr/bin/hautomo-client "https://github.com/function61/hautomo/releases/download/..."

	return 0
}

function installTurboBob {
	downloadAndInstallSingleBinaryProgram /usr/bin/bob "https://github.com/function61/turbobob/releases/download/20200910_1241_90ec91c9/bob_linux-amd64"

	joonas-os-ram-image
}

function installJames {
	downloadAndInstallSingleBinaryProgram /usr/bin/james "https://bintray.com/function61/dl/download_file?file_path=james%2F20190628_1117_a7803323%2Fjames_linux-amd64"
}

function installAllTheBits {
	# downloadAndInstallSingleBinaryProgram /usr/bin/atb "https://bintray.com/joonas-fi/atb/..."

	return 0
}

function fixPermissions {
	# for some reason root ends up with other:write. TODO: find out why
	chmod o-w /

	# any file we wrote, we wrote as root
	chown -R "$username:$username" "/home/$username"
}

function step {
	local name="$1"

	mkdir -p /tmp/.joonas-os-install

	local flagFileCompleted="/tmp/.joonas-os-install/${name}.flag-completed"

	echo "# $name"

	if [ -e "$flagFileCompleted" ]; then
		echo "Already run successfully; skiping"
		return 0
	fi

	"$name"

	touch "$flagFileCompleted"
}

function installationProcess {
	step configureFilesystemTable

	step setupNetwork

	step aptUpdateAndUpgrade

	step aptEnableAddAptRepository

	step installLanguagePack

	step createUser

	step overlayOverrides

	step configureTimezoneAndLocale

	# step installSnapd

	# needs to be installed before kernel (LVM modules need to be present in initrd I guess)
	# we don't use it right yet, but better have it ready
	# step installLvm

	step installKernelAndBootloader

	step setupRootOverlay

	step installGpuDriver

	step installFavouriteBaseUtils

	step installGit

	step installMercurial

	step installLazygit

	step installDocker

	step installQemu

	step installLf

	step installGraphicalEnvironment

	step installPeek

	step installAutorandr

	step installGraphicalSessionManager

	step installI3Gaps

	step installSublimeText

	step installFirefox

	step installLibreoffice

	step installCroc

	step installVarasto

	step installHautomoClient

	step installTurboBob

	step installJames

	step installAllTheBits

	step fixPermissions
}

installationProcess
