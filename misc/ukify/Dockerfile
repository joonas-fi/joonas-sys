FROM ubuntu:noble

WORKDIR /workspace

ENTRYPOINT ["/usr/bin/ukify"]

RUN apt update && DEBIAN_FRONTEND=noninteractive apt-get install -y systemd-ukify systemd-boot-efi

