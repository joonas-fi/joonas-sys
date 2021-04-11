
alias ..="cd .."
alias ...="cd ../.."
alias ....="cd ../../.."
alias .....="cd ../../../.."

tailscale() {
	docker exec -it tailscale_tailscale_1 tailscale "$@"
}

exiftool() {
	docker run --rm -v "$(pwd):/workspace" joonas/exiftool "$@"
}

convert() {
	docker run --rm -v "$(pwd):/workspace" joonas/imagemagick convert "$@"
}

croc() {
	docker run -it --rm -v "$(pwd):/workspace" joonas/croc "$@"
}

nslookup() {
	docker run -it --rm alpine nslookup "$@"
}

dig() {
	docker run -it --rm joonas/dig "$@"
}

htop() {
	docker run -it --rm --pid=host joonas/htop "$@"
}

iotop() {
	docker run --rm -it --privileged --net=host --pid=host joonas/iotop "$@"
}

nethogs() {
	docker run --rm -it --net=host --pid=host --privileged joonas/nethogs "$@"
}

pstree() {
	docker run --rm -it --pid=host joonas/psmisc pstree "$@"
}

killall() {
	docker run --rm -it --pid=host --privileged joonas/psmisc killall "$@"
}

lspci() {
	docker run --rm -it joonas/lspci "$@"
}

ncdu() {
	# need to pass locale for non-ASCII chars to work

	if [[ $# -eq 0 ]]; then
		docker run --rm -it -v "/:/host:ro" -e LANG joonas/ncdu "$(pwd)"
	else
		docker run --rm -it -v "/:/host:ro" -e LANG joonas/ncdu "$@"
	fi
}
