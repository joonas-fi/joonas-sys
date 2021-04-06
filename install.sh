#!/bin/bash -eu

repodir="/tmp/repo"

source "${repodir}/secrets.env"

username="joonas"

# for how to generate https://askubuntu.com/a/667842
userPasswordHash="$USER_PASSWORD_HASH" # from secrets.env

# This installation process outline was shaped by:
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

	# the .network files were specified in overrides/
	# we could do network config with /etc/network but I guess systemd-networkd has advantages?

	# https://wiki.archlinux.org/index.php/systemd-networkd
	# for some reason the network configuration daemon is not up by default
	systemctl enable systemd-networkd
}

function reconfigureTzdata {
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

# LVM is used for overlaying virtual block devices on top of physical disks to achieve encryption,
# extending disks dynamically etc.
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

# programmatic steps related to having / be overlayfs that redirects writes to /persist/sys_N_diffs
#
# most of the important things are already done in scripts in our overrides/etc/initramfs-tools/
function setupRootOverlay {
	# instruct pre-boot environment to have overlay kernel module loaded
	# we could have this as static file, but then upstream changes would get overwritten
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
		--password "$userPasswordHash" \
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

	# any file we wrote under user's home dir, we wrote as root
	chown -R "$username:$username" "/home/$username"

	# for some reason root ends up with other:write. was it because rsync?
	chmod o-w /
}

# snap/snapcraft is Docker-like but mainly focused for GUI apps
function installSnapd {
	apt install -y snapd
}

function installFavouriteBaseUtils {
	# bsdmainutils = hexdump
	# usbutils = lsusb
	# pciutils = lspci
	# dnsutils = nslookup, dig
	# imagemagick = convert
	apt install -y \
		htop \
		iotop \
		curl \
		wget \
		unzip \
		jq \
		pv \
		ncdu \
		imagemagick \
		vim \
		strace \
		pciutils \
		usbutils \
		bsdmainutils \
		dnsutils \
		nmap \
		exiftool \
		tree
}

function installPrinterDriver {
	# we seem to have CUPS etc. installed already :O
	# Epson AcuLaser M1400
	apt install -y printer-driver-foo2zjs
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
		curl -fsSL https://bootstrap.pypa.io/pip/2.7/get-pip.py --output get-pip.py
		python2 get-pip.py
		rm -rf "$tempInstallDir"
	)

	# https://pypi.org/project/hg-git/
	# (clone from github didn't work without these additional modules)
	pip install hg-git brotli ipaddress
}

# fantastic CLI UI for Git
function installLazygit {
	curl -fsSL "https://github.com/jesseduffield/lazygit/releases/download/v0.23.7/lazygit_0.23.7_Linux_x86_64.tar.gz" \
		| tar -C /usr/bin -xz lazygit
}

function installDocker {
	# don't install recommends, because it installs "Ubuntu fan". bring in most of recommended manually
	apt install --no-install-recommends -y docker.io git cgroupfs-mount pigz xz-utils

	apt install -y docker-compose

	# add user to Docker group, so we don't need to "$ sudo ..." all docker commands
	usermod -aG docker "$username"

	# Docker is not enabled by default.
	# (instead it is socket-activated, i.e. "$ docker ps" would start "always-up" services)
	systemctl enable docker
}

# Virtual Machines
function installQemu {
	apt install -y qemu-system-x86
}

# CLI filesystem navigator
function installLf {
	curl -fsSL https://github.com/gokcehan/lf/releases/download/r19/lf-linux-amd64.tar.gz \
		| tar -C /usr/bin/ -xzf -
}

function installGraphicalEnvironment {
	# (why "noninteractive"? read rant about gdm3 below)
	#
	# unless we specify slick-greeter, it's going to default to unity-greeter which pulls in
	# parts of Unity also

	# - lightdm is a display manager (= GUI for logging in to your desktop)
	# - rofi is an application launcher
	# - xwallpaper might not be always required (once hautomo-client can set wallpapers without it)
	# - ttf-ancient-fonts because emojis didn't render (https://www.omgubuntu.co.uk/2014/11/see-install-use-emoji-symbols-ubuntu-linux)
	# - fonts-noto-color-emoji to get colored emojis for i3 workspace symbols
	DEBIAN_FRONTEND=noninteractive apt install -y \
		xfce4 \
		xfce4-screensaver \
		xfce4-notifyd \
		lightdm \
		slick-greeter \
		alsa \
		compton \
		xwallpaper \
		rofi \
		mousepad \
		vlc \
		ttf-ancient-fonts \
		fonts-firacode \
		fonts-noto-color-emoji \
		fonts-hack \
		fonts-powerline

	# before for some reason lightdm had to be installed in a separate "$ apt" call, otherwise gdm3
	# would get pulled in. now after I removed something from here it gets pulled in anyway. JFC I
	# tried to research which package pulls it, I couldn't figure it out and now we have conflict
	# because we have 2 managers, and now we've to fix it. UGH!

	echo /usr/bin/lightdm > /etc/X11/default-display-manager

	rm /etc/systemd/system/display-manager.service

	ln -s /lib/systemd/system/lightdm.service /etc/systemd/system/display-manager.service
}

# a clipboard manager is required if you don't want clipboard to empty when the program exits where
# you copied the data from
function installClipboardManager {
	curl -fsSL "https://github.com/xrelkd/clipcat/releases/download/v0.5.0/clipcat-v0.5.0-x86_64-unknown-linux-gnu.tar.gz" \
		| tar -xz -C /usr/bin -f - clipcat-menu clipcat-notify clipcatctl clipcatd
}

function installTerminalEmulator {
	# for alacritty
	add-apt-repository ppa:aslatter/ppa

	apt install -y alacritty
}

# screenshots with annotation support (= draw arrows etc.)
function installFlameshot {
	apt install -y flameshot
}

# a screen recorder (.gif, .mp4, ..)
function installPeek {
	apt install -y peek

	# 'dbus-launch --exit-with-session' prefix: https://askubuntu.com/a/311988
	su "$username" -c "dbus-launch --exit-with-session gsettings set com.uploadedlobster.peek persist-save-folder /tmp"
	su "$username" -c "dbus-launch --exit-with-session gsettings set com.uploadedlobster.peek recording-framerate 24"
}

# When monitors connect/disconnects, set appropriate screen configuration automatically
function installAutorandr {
	apt install -y autorandr
}

# tiling window manager (gaps fork for prettier visuals)
function installI3Gaps {
	# status bar for i3. recommended would install dzen2 and i3-wm
	# apt install --no-install-recommends -y i3status

	# i3-gaps (a fork of i3 with gaps support) exists in speed-ricer repo
	add-apt-repository -y ppa:kgilmer/speed-ricer

	# recommended would install dunst.
	# suckless-tools = dmenu (used for workspace renaming)
	# session means setting file so the i3 session shows up in greeter
	apt install --no-install-recommends -y i3-gaps-wm i3-gaps-session i3status suckless-tools
}

# Text editor
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

# "Excel and Word"
function installLibreoffice {
	apt install -y libreoffice-calc libreoffice-writer
}

# P2P file transfer program, it's really magical! Go-port of "Magic Wormhole"
function installCroc {
	(
		local tempInstallDir="/tmp/croc-install"

		mkdir "$tempInstallDir" && cd "$tempInstallDir"
		curl -fsSL -o croc.deb https://github.com/schollz/croc/releases/download/v8.6.7/croc_8.6.7_Linux-64bit.deb
		dpkg -i croc.deb 
		rm -rf "$tempInstallDir"
	)
}

function installWireguard {
	apt install -y wireguard

	# installation unfortunately replaces our symlink with an empty dir
	if [ ! -L /etc/wireguard ]; then # not a symlink? => replace our symlink back
		rm -rf /etc/wireguard
		cp -a "${repodir}/overrides/etc/wireguard" /etc/wireguard
	fi
}

function installWireshark {
	# it asks "Should non-superusers be able to capture packets?"
	DEBIAN_FRONTEND=noninteractive apt install -y wireshark
}

# Centralized file storage
function installVarasto {
	downloadAndInstallSingleBinaryProgram /usr/bin/sto "https://github.com/function61/varasto/releases/download/20200626_1423_4cd3ecf8/sto_linux-amd64"
}

# Integrate home automation to my PC
function installHautomoClient {
	# downloadAndInstallSingleBinaryProgram /usr/bin/hautomo-client "https://github.com/function61/hautomo/releases/download/..."

	return 0
}

# Development environment/build tool
function installTurboBob {
	downloadAndInstallSingleBinaryProgram /usr/bin/bob "https://function61.com/go/turbobob-latest-linux-amd64"
}

# easy provisioning of development SSL certs
function installMkcert {
	# dependency of mkcert
	apt install -y libnss3-tools

	downloadAndInstallSingleBinaryProgram /usr/bin/mkcert "https://github.com/FiloSottile/mkcert/releases/download/v1.4.3/mkcert-v1.4.3-linux-amd64"
}

# Docker cluster management etc.
function installJames {
	ln -s /persist/work/james/rel/james_linux-amd64 /usr/bin/james
	# downloadAndInstallSingleBinaryProgram /usr/bin/james "https://bintray.com/function61/dl/download_file?file_path=james%2F20190628_1117_a7803323%2Fjames_linux-amd64"
}

# Podcast, Youtube etc. hoarding
function installAllTheBits {
	# downloadAndInstallSingleBinaryProgram /usr/bin/atb "https://bintray.com/joonas-fi/atb/..."

	return 0
}

function disableUnnecessaryBloat {
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

	step reconfigureTzdata

	step installSnapd

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

	step installClipboardManager

	step installPrinterDriver

	step installTerminalEmulator

	step installFlameshot

	step installPeek

	step installAutorandr

	step installI3Gaps

	step installSublimeText

	step installFirefox

	step installLibreoffice

	step installCroc

	step installWireguard

	step installWireshark

	step installVarasto

	step installHautomoClient

	step installTurboBob

	step installMkcert

	step installJames

	step installAllTheBits

	step disableUnnecessaryBloat
}

installationProcess
