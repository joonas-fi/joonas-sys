# TODO: add version here
FROM ubuntu

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
CMD /repo/bootstrap.sh

ADD install.sh bootstrap.sh /repo/
ADD overrides/ /repo/overrides/
