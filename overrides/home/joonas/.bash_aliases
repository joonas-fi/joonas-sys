
# starship.rs integration
eval "$(starship init bash)"

# zoxide integration
eval "$(zoxide init bash)"

# McFly integration
eval "$(mcfly init bash)"

# make "$ cd project" equivalent to "$ cd ~/work/project"
# must NOT be exported, see https://bosker.wordpress.com/2012/02/12/bash-scripters-beware-of-the-cdpath/
CDPATH=.:~/work/

alias ..="cd .."
alias ...="cd ../.."
alias ....="cd ../../.."
alias .....="cd ../../../.."

# e as in "edit in favourite editor")
alias e="hx"

# mnemonic: "cd directory"
alias cdd="cd \$(find . -type d | fzf)"

alias less="less --chop-long-lines --raw-control-chars"

alias pubkey="ssh-keygen -y -f ~/.ssh/id_rsa"

# horizontal rule
alias hr="jsys hr"

# when mapping full host FS to inside container like under previx /host/... we can't set workdir
# necessarily as-is (because absolute symlinks like /etc/foo.conf are not resolvable from different
# perspective as it would need to be /host/etc/foo.conf now), but instead evaluate symlinks as workaround.
_pwd_symlinksresolved() {
	# like "$ pwd", but with symlinks evaluated to their final destination
	readlink --canonicalize-existing .
}

lazygit() {
	local repoName="$(basename "`pwd`")"
	local terminalTitle="lazygit $repoName"

	# set terminal title
	printf "\e]2;$terminalTitle\a"

	# to not invoke our alias again
	/usr/bin/lazygit

	# don't know better way to reset, empty or space doesn't seem to work
	printf "\e]2;-\a"
}

alpine() {
	docker run --rm -it --network=host -v "$(pwd):/workspace" --workdir /workspace alpine "$@"
}

exiftool() {
	docker run --rm --net=none -v "$(pwd):/workspace" joonas/exiftool "$@"
}

convert() {
	docker run --rm --net=none -v "$(pwd):/workspace" joonas/imagemagick convert "$@"
}

croc() {
	docker run --rm -it -v "$(pwd):/workspace" joonas/croc "$@"
}

nslookup() {
	docker run --rm -it alpine nslookup "$@"
}

dig() {
	docker run --rm -it joonas/dig "$@"
}

htop() {
	# /etc/passwd to resolve user ids to usernames
	docker run --rm -it \
		--net=none \
		--pid=host \
		-v ~/.config/htop/htoprc:/etc/htoprc \
		-v /etc/passwd:/etc/passwd:ro \
		joonas/htop "$@"
}

iotop() {
	docker run --rm -it --net=host --privileged --pid=host joonas/iotop "$@"
}

nethogs() {
	docker run --rm -it --net=host --pid=host --privileged joonas/nethogs "$@"
}

ethtool() {
	docker run --rm -it --net=host ghcr.io/r-xs-fi/ethtool "$@"
}

pstree() {
	docker run --rm -it --net=none --pid=host joonas/psmisc pstree "$@"
}

killall() {
	docker run --rm -it --net=none --pid=host --privileged joonas/psmisc killall "$@"
}

lspci() {
	docker run --rm -it --net=none ghcr.io/r-xs-fi/lspci "$@"
}

fdisk() {
	docker run --rm -it --net=none --privileged joonas/fdisk "$@"
}

gdisk() {
	docker run --rm -it --net=none --privileged --entrypoint=/usr/bin/gdisk joonas/fdisk "$@"
}

hollywood() {
	# jess/hollywood does not work :(
	docker run --rm -it --net=none bcbcarl/hollywood "$@"
}

asciiquarium() {
	docker run --rm -it --net=none danielkraic/asciiquarium "$@"
}

radeontop() {
	docker run --rm -it --net=none --privileged joonas/radeontop "$@"
}

pwdgen() {
	docker run --rm -it --net=none joonas/pwdgen:20211113_1928_0fea588c "$@"
}

nmap() {
	docker run --rm -it --net=host joonas/nmap "$@"
}

stress-ng() {
	docker run --rm -it joonas/stress-ng "$@"
}

netstat() {
	docker run --rm -t --net=host alpine netstat "$@"
}

arp() {
	docker run --rm -it --net=host alpine arp "$@"
}

iperf() {
	docker run --rm -it --net=host joonas/iperf "$@"
}

file() {
	docker run --rm -it --net=none -v "/:/host:ro" --workdir="/host$(_pwd_symlinksresolved)" joonas/file "$@"
}

rpm2cpio() {
	docker run --rm -t --net=none -v "/:/host:ro" --workdir="/host$(_pwd_symlinksresolved)" joonas/rpm2cpio "$@"
}

smartctl() {
	docker run --rm -it --net=none --privileged joonas/smartmontools "$@"
}

qrencode() {
	docker run --rm -t --net=none joonas/qrencode "$@"
}

figlet() {
	docker run --rm -i --net=none joonas/figlet "$@"
}

ncdu() {
	# need to pass locale for non-ASCII chars to work

	if [[ $# -eq 0 ]]; then
		docker run --rm --net=none -it -v "$(pwd):/workspace" ghcr.io/r-xs-fi/ncdu
		# docker run --rm --net=none -it -v "/:/host:ro" -e LANG joonas/ncdu "/host$(_pwd_symlinksresolved)"
	else
		docker run --rm --net=none -it -v "/:/host:ro" -e LANG joonas/ncdu "$@"
	fi
}

awscli() {
	docker run --rm -it -v "$(pwd):/aws" --entrypoint= amazon/aws-cli bash
}

hey() {
	# host networking for correct perspective to "localhost" and to minimize perf impact
	# (am not sure how packets from container virtual NICs are routed)
	docker run --rm -it --net=host joonas/hey "$@"
}

lshw() {
	# privileges needed for /dev, /sys etc access
	docker run --rm -it --net=host --privileged joonas/lshw "$@"
}

sqlite() {
	docker run --rm -it --net=none -v "$(pwd):/workspace" --workdir=/workspace joonas/sqlite "$@"
}

pdfcpu() {
	docker run --rm -it --net=none -v "$(pwd):/workspace" joonas/pdfcpu "$@"
}

uuidgen() {
	docker run --rm -it --net=none joonas/uuidgen "$@"
}

cmatrix() {
	docker run --rm -it --net=none joonas/cmatrix "$@"
}

telnet() {
	docker run --rm -it joonas/telnet "$@"
}

whois() {
	docker run --rm -it ghcr.io/r-xs-fi/whois "$@"
}

mapscii() {
	docker run --rm -it joonas/telnet mapscii.me
}

cowsay() {
	docker run --rm -it --net=none joonas/cowsay "$@"
}

doge() {
	docker run --rm -it --net=none joonas/doge "$@"
}

tokei() {
	docker run --rm -it --net=none -v "$(pwd):/workspace" joonas/tokei "$@"
}

neofetch() {
	docker run --rm -it --network=host -v /etc/os-release:/etc/os-release:ro joonas/neofetch "$@"
}

ffmpeg() {
	docker run --rm -it --net=none -v "$(pwd):/workspace" ghcr.io/r-xs-fi/ffmpeg "$@"
}

yulelog() {
	docker run --rm -it --net=none joonas/yulelog "$@"
}

pipes.sh() {
	docker run --rm -it --net=none joonas/pipes.sh "$@"
}

go-life() {
	docker run --rm -it --net=none joonas/go-life "$@"
}

lolcat() {
	# --force -> Force color even when stdout is not a tty
	docker run --rm -i --net=none joonas/lolcat --force "$@"
}

scooter() {
	docker run --rm -it --net=none -v "$(pwd):/workspace" --user=$(id -u):$(id -g) ghcr.io/r-xs-fi/scooter:latest "$@"
}

