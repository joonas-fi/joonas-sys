# using latest LTS. they are released every two years:
# https://wiki.ubuntu.com/Releases
FROM ubuntu:oracular

# debootstrap bootstraps (= minimal base with APT so you can build from there) another
# file tree for installing Ubuntu inside it. Basically from an existing Ubuntu installation you can
# "bootstrap" /example-installation and install a working system there inside a chroot.
RUN apt update && apt install -y debootstrap

# cache tree bootstrap, so we benefit from Docker's cache if we have to run this process multiple
# times (we don't have to download & extract packages multiple times)
#
# https://help.ubuntu.com/community/Installation/FromLinux#Debootstrap
#
# use same release as the Docker image we're using (os-release gives us $VERSION_CODENAME)
RUN mkdir /debootstrap-cache && . /etc/os-release && debootstrap "$VERSION_CODENAME" /debootstrap-cache

# - copies debootstrap-cache into target tree
# - begins installation by calling install.sh inside chroot
ENTRYPOINT ["/repo/bin/run-step-in-container.sh"]

WORKDIR /repo/install-steps

# ADD bin/run-step-in-container.sh /repo/bin/run-step-in-container.sh
# ADD overrides/ /repo/overrides/
# ADD install-steps/ /repo/install-steps/
