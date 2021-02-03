This is my personal OS of choice (currently Ubuntu) & system installation (programs & conf I use) as
a code. You might be familiar with "dotfiles" - this is a bit further. :) (Because I'm a giant nerd.)

This is not a generic operating system ("distro") that could help you - it's my personalized system,
i.e. difference between an OS and an OS installation after one has set it up for her liking. But I'm
sharing this for other people to get inspired & share ideas!

In summary, running these scripts produces:

- An image for system partition with up-to-date Ubuntu installation + apps I use + config I use
- An image of boot partition

These two image files can be taken to a totally new computer, boot it with a USB stick containing the
images and use the utility in the USB stick to write the image partitions to the disk and have the
system exactly the way I'm used to using it!


Contents:

- [How does it work](#how-does-it-work)
  * [Partitioning](#partitioning)
  * [No updates to the system](#no-updates-to-the-system)
  * [Installation, configuration, repository file layout](#installation-configuration-repository-file-layout)
  * [Handling state](#handling-state)
- [Why this approach?](#why-this-approach-)
- [Long-term goals](#long-term-goals)
- [How to use](#how-to-use)
  * [Process](#process)
  * [Scripts](#scripts)
  * [Build environment portability](#build-environment-portability)
- [Road to memory safety](#road-to-memory-safety)
- [Additional reading](#additional-reading)


How does it work
----------------

Summary:

- My system image (OS + installed apps) is immutable
- Despite immutability testing new software is easy (just `$ apt install <program>` like always)
	* Short-term root state changes (like installing a new program) are redirected to a "diff" tree
	  using [overlayfs](https://wiki.archlinux.org/index.php/Overlay_filesystem) so I don't
	  accidentally lose short-term data after reboot. But short-term data is wiped out weekly on purpose.
	* It is easy for me to audit that I'm not accidentally throwing away important data because I can
	  inspect the diff tree.
* No automatic updates to software, but move to a fresh install of up-to-date system weekly.
* All configuration comes from this repo ("system state as code").


### Partitioning

Overview as a drawing:

![](docs/overview.png)

Same in text:

It might be easiest to begin by explaining my partition layout:

```
sda           8:0    0 894.3G  0 disk
├─sda1        8:1    0   256M  0 part /boot/efi
├─sda2        8:2    0  47.7G  0 part
├─sda3        8:3    0  47.7G  0 part /persist/sys-current-rom
└─sda4        8:4    0 798.8G  0 part /persist
```

First partition is UEFI system partition (AKA "ESP") - the bootloader (which is responsible for
starting the OS). `sda2` and `sda3` are equal sized active/passive **readonly** system images.

The persist partition is actually important data (work files, settings etc.), i.e. it doesn't contain
my installed programs or anything a random program decides to write somewhere, even in my `/home`
directory.


### No updates to the system

I disable updates in my OS and the programs (Firefox etc.), but I run these steps weekly:

1. Build & flash a freshly-installed kernel+drivers+programs into the passive partition
2. (Optionally) Test the new passive partition's system in a VM
3. Switch roles of active-passive partitions (the old passive becomes the new active)
4. Reboot into the new active partition

As a result:

- Updates won't ever break my running system all of a sudden
	* Because running system is never updated
	* Updates are important, but I achieve the same by just starting each week with a freshly
	  installed system containing newest packages
	* Semantically my software never gets updated. Software is much simpler if you don't ever have
	  to worry about update logic (or removal for that matter). But you have to get good at
	  identifying & managing state!
- I get to decide the exact time when I apply all the updates in an atomic manner
- If updates break anything, I can rollback by booting into the previous week's system that worked


### Installation, configuration, repository file layout

The most interesting file is [install.sh](install.sh). It specifies all the packages to install.

Second most interesting directory tree is [overrides/](overrides/), which contains all customized
files I want to be present in the image:

- Some software runs with its default configuration (no overrides needed)
- For some software I need to override their files (or provide additional ones)
	* [Firefox is one such example](overrides/usr/lib/firefox/)

The `Dockerfile` is mainly about getting this to build anywhere (think: build this Debian-based image
from e.g. Arch Linux) with minimal dependencies.

`bootstrap.sh` mainly bootstraps Debian installation environment and pivots the chroot inside the
 to our system tree that we're building and calls `install.sh` where the actual installation can
 begin after having a working package manager.


### Handling state

There are roughly three categories of data:

| Type of file | Source of data | Stored in | Example |
|--------------|--------|-----------|---------|
| Static file installed by an application | OS package manager | System image | Executable or config file that you didn't change |
| File that only needs to be changed rarely | This repository | System image | Application's config file that you customized, like [Firefox customizations](overrides/usr/lib/firefox/browser/defaults/preferences/user.js) |
| Persistent data, state that changes often | User generated important data | `/persist` partition | Application's state files, photographs you took etc. |

Application's state files, even whole directory trees can be redirected to `/persist` partition by
symlinking its directory or a file. For example I want to persist Docker's state so my
`/var/lib/docker` is symlinked to `/persist/docker-data`.
[It's specified in this repo](overrides/var/lib/docker).

There's one special case for secret files, like your SSH private keys or other sensitive data, which
basically is rarely-changing data (therefore could be stored in repo), but for security reasons
shouldn't be stored in the repo. In this case I can make the file in  `/home/joonas/...` be a symlink to
`/persist/...`
([see example](https://github.com/joonas-fi/joonas-sys/blob/ab68d9e47612ffb8984c37343e21f091e1599445/overrides/home/joonas/.ssh/id_rsa)),
so I can manage the state outside of the repo without having to configure the software to look for the
file from my special location.


Why this approach?
------------------

This might seem like much added complexity to you, and that's a fair argument. But I'm thinking I'm
just paying the price beforehand, because it's easier to do it now than later ("Leaning in to the
pain"). Let me explain..

The thing that has always bothered me is that computers tend to intermingle (or at least make it too
easy to):

- Interesting state (that is worth preserving, backing up) and
- Totally **un**interesting state

This makes moving between systems hard (think: migrating to a new computer or mobile device).

As a Windows user I used to obsess over how the data was laid out in my `C:` and `D:` drives. I used
to get angry if some badly behaving software wrote its data or log files to the root of the partition
or make a folder directly under the root:

![](docs/windows-c-drive-unnecessary-crap.png)

Linux is not immune to this. Here's how my freshly installed system looks:

```
/home/joonas/
├── .Xauthority
├── .bash_history
├── .bash_logout
├── .bashrc
├── .cache
├── .config
├── .dbus
├── .dmrc
├── .gnupg
├── .lesshst
├── .local
├── .mozilla
├── .mplayer
├── .profile
├── .selected_editor
├── .ssh
├── .sudo_as_admin_successful
├── .thunderbird
├── .vim
├── .viminfo
├── .wget-hsts
├── .wine
├── .xsession-errors
├── Desktop
├── Downloads
├── snap
└── work -> /persist/work
```

The only entries I placed there myself was `.config/`, `.ssh/` and `work/`. Most of `./config/` is also
filled up with stuff I didn't put there.

Without the drastic approach I'm taking, I don't think there is other way to manage one's system
state in a way that doesn't leave you with dread on data loss (did I backup everything I care about?).

You can of course backup your entire system but then you're left with countless unnecessary files
that you've to keep forever unless you take the time to dig into the backup to inspect if there were
interesting files to recover before deleting the long-gone system.

Would you say identifying interesting state would be easier to do now (or at most a week after the state
was created), than to leave it for you do do ten years from now?

[Graham Christensen](https://grahamc.com/blog/erase-your-darlings) put it eloquently:

> Over time, a system collects state on its root partition. This state lives in assorted directories
> like /etc and /var, and represents every under-documented or out-of-order step in bringing up the services.
>
>   “Right, run myapp-init.”
>
> These small, inconsequential “oh, oops” steps are the pieces that get lost and don’t appear in your runbooks.
>
>   “Just download ca-certificates to … to fix …”
>
> Each of these quick fixes leaves you doomed to repeat history in three years when you’re finally
> doing that dreaded RHEL 7 to RHEL 8 upgrade.


Long-term goals
---------------

To serve me well:

- Solve state management
- Reduce complexity - there's tremendous complexity in updating software:
	* Microsoft's MSI
	* Linux's initramfs hooks
	* boot partition versioned kernels and initrd's
	* It all just gets simpler if software doesn't need update/remove capabilities.
- Minimal operating system itself, move as much to containers like Docker or Snapd (or minimal-dependency binaries like Go can produce)
- Get rid of as much C/C++ (= memory unsafe) code as possible to increase security


How to use
----------

How I use these tools to manage my system.


### Process

![](docs/process.png)

NOTE: there is no hard requirement to build the systree in RAM, but I do it so:

- I don't end up writing to my passive partition if the build process fails (wrecks the previously
  good partition)
- Reduce unnecessary I/O - after successful build I [rsync](https://en.wikipedia.org/wiki/Rsync) only
  the difference from RAM to the passive partition and thus reduce SSD wear.
- Sure I could use spare space in `/persist` to host a temporary loopback partition to circumvent the
  failed build problem, but I have enough RAM, RAM is faster and it reduces unnecessary disk I/O.


### Scripts

WARNING: these are not safe to run unless you understand what they do first. Some scripts write to
partitions, some scripts modify your boot partition.. Again, I'm sharing these for education use, not
as safe usable programs for anyone else!

The whole process centers around these:

``` console
$ bin/build-systree.sh
$ bin/systree-to-raw-image.sh
$ bin/test-in-vm.sh
```

(NOTE: `$ sudo` probaby required.)


### Build environment portability

The above build process should work on pretty much any distro, you only need Docker installed
(, and qemu if you want to test the built system in a VM).


Road to memory safety
---------------------

In the graphical stack:

| Component | Memory safe | Program | Notes |
|-----------|-------------|---------|-------|
| Display server | | Xorg | |
| Display manager | | LightDM | |
| Greeter | | lightdm-gtk-greeter | |
| Window manager |  | i3 | [Investigate memory safe alternatives](https://users.rust-lang.org/t/is-there-a-tiling-window-manager-for-linux-that-is-written-and-configurable-in-rust/4407)9sublim |
| Compositor |  | compton | |
| Screensaver |  | xfce4-screensaver | |
| Screenshot app |  | xfce4-screenshooter | |
| Notification daemon |  | dunst | |
| Clipboard manager | | xfce4-clipman | |
| Terminal | | xfce4-terminal (switch to alacritty?) | |
| Program launcher | | rofi | |
| Display settings manager | python? | autorandr | |
| Media player control | ✓ | Hautomo's playerctl | |


Additional reading
------------------

- [Erase your darlings](https://grahamc.com/blog/erase-your-darlings) (powerful thought: "Leaning in to the pain")

