
alias ..="cd .."
alias ...="cd ../.."
alias ....="cd ../../.."
alias .....="cd ../../../.."

alias less="less --chop-long-lines --raw-control-chars"

tailscale() {
	docker exec -it tailscale_tailscale_1 tailscale "$@"
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

pstree() {
	docker run --rm -it --net=none --pid=host joonas/psmisc pstree "$@"
}

killall() {
	docker run --rm -it --net=none --pid=host --privileged joonas/psmisc killall "$@"
}

lspci() {
	docker run --rm -it --net=none joonas/lspci "$@"
}

fdisk() {
	docker run --rm -it --net=none --privileged joonas/fdisk "$@"
}

gdisk() {
	docker run --rm -it --net=none --privileged --entrypoint=/usr/bin/gdisk joonas/fdisk "$@"
}

radeontop() {
	docker run --rm -it --net=none --privileged joonas/radeontop "$@"
}

nmap() {
	docker run --rm -it --net=host joonas/nmap "$@"
}

stress-ng() {
	docker run --rm -it joonas/stress-ng "$@"
}

netstat() {
	docker run --rm -it --net=host alpine netstat "$@"
}

file() {
	docker run --rm -it --net=none -v "/:/host:ro" --workdir="/host$(pwd)" joonas/file "$@"
}

ncdu() {
	# need to pass locale for non-ASCII chars to work

	if [[ $# -eq 0 ]]; then
		docker run --rm --net=none -it -v "/:/host:ro" -e LANG joonas/ncdu "$(pwd)"
	else
		docker run --rm --net=none -it -v "/:/host:ro" -e LANG joonas/ncdu "$@"
	fi
}

awscli() {
	docker run --rm -it -v "$(pwd):/aws" --entrypoint= amazon/aws-cli bash
}
