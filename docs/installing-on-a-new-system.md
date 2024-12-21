Installing on a new system
==========================

## Overview

- Install regular Ubuntu on the system
- Use the tools provided by Ubuntu to "Infect" the Ubuntu with joonas-sys and then boot into it

## Install regular Ubuntu on the system

no instructions needed here

## Install joonas-sys

### Copy jsys binary

Copy `$ jsys` to the system


### Create the directory structures

Make the running Ubuntu system fake it has a sysroot mount point pointing to the final sysroot even though there isn't one yet.

It should look like this:

```
/sysroot
├── apps -> /apps
└── lost+found -> /lost+found
```


```shell
sudo mkdir /sysroot
sudo mkdir /apps
sudo ln -s /apps /sysroot/apps
sudo ln -s /lost+found /sysroot/lost+found
```

Then start creating the important "apps":

```shell
sudo mkdir -p /sysroot/apps/SYSTEM/{backlight-state,rfkill-state,lowdiskspace-check-rules}

sudo mkdir -p /sysroot/apps/{work,zoxide,flatpak,flatpak-appdata,netplan,mcfly,git\ config,Desktop,ssh-server,OS-diff-work}

sudo ssh-keygen -t ed25519 -f /sysroot/apps/ssh-server/ssh_host_ed25519_key -N ""

sudo mkdir -p /sysroot/apps/{varasto/varasto-work,docker/config,docker/data}

sudo cp /etc/hostname /sysroot/apps/SYSTEM/hostname
sudo cp /etc/machine-id /sysroot/apps/SYSTEM/machine-id

# need to be user-writable
sudo chown 1000:1000 /sysroot/apps/{zoxide,mcfly,flatpak-appdata}

sudo touch /sysroot/swapfile  # FIXME: invalid, not created with mkswap

sudo jsys lowdiskspace-checker rule-create root /
sudo jsys lowdiskspace-checker rule-set-threshold --gb 20 root

sudo curl -fsSL -o /sysroot/apps/SYSTEM/background.png https://github.com/user-attachments/assets/0d22f401-d4be-4d23-89ea-85d1ef789815

sudo mkdir /sysroot/apps/{OS-checkout,OS-diff,OS-repo}

sudo setfattr -n user.xdg.robots.backup -v false /sysroot/lost+found /sysroot/apps/{SYSTEM,OS-checkout,OS-diff,OS-repo,docker/data,flatpak,flatpak-appdata}

```


### Make bootable

```shell
cd /sysroot/apps/OS-repo
sudo ostree init --repo=. && sudo ostree remote add --repo=. --no-gpg-verify fi.joonas.os https://fi-joonas-os.ams3.digitaloceanspaces.com/ostree/
sudo jsys ostree pull
sudo jsys flash efi
sudo cp /tmp/ukifybuild/BOOTx64.efi /boot/efi/EFI/BOOT/BOOTx64.efi
```

Ensure that the root partition's filesystem label is the one the cmdline looks by:

```shell
sudo tune2fs -L persist /dev/nvme...
```

### After first boot

Ensure that it reaches internet.

Clean up other dirs than `apps` (or `lost+found`) from `/sysroot` as the Ubuntu is no longer needed.

Log in to Tailscale:

```shell
sudo tailscale up
```

Ensure function22 is running.

Fix Flatpak if needed:

```shell
sudo flatpak remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
sudo chown $(id -u):$(id -g) ~/.var/app
```

Get access to web:

```shell
 sudo flatpak install --assumeyes flathub org.mozilla.firefox
```

Configure Varasto:

- Go create new auth token

Run:

```shell
sto config init https://$HOSTNAME $TOKEN /sto
# because Varasto doesn't write to symlink target but replaces symlink, one has to fix this
sudo mv ~/.config/varasto/client-config.json /sysroot/apps/varasto/client-config.json
ln -s /sysroot/apps/varasto/client-config.json ~/.config/varasto/client-config.json
```
