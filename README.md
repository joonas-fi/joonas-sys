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
- [Long-term goals](#long-term-goals)
- [How to use](#how-to-use)
  * [Process](#process)
  * [Scripts](#scripts)
  * [Build environment portability](#build-environment-portability)
- [Road to memory safety](#road-to-memory-safety)
- [Additional reading](#additional-reading)


How does it work
----------------

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

### Process

![](docs/process.png)


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

| Component | Memory safe | Program |
|-----------|-------------|---------|
| Display server | | Xorg |
| Display manager | | LightDM |
| Greeter | | lightdm-gtk-greeter |
| Window manager |  | i3 |
| Compositor |  | compton |
| Screensaver |  | xfce4-screensaver |
| Screenshot app |  | xfce4-screenshooter |
| Notification daemon |  | dunst |
| Clipboard manager | | xfce4-clipman |
| Terminal | | xfce4-terminal (switch to alacritty?) |
| Program launcher | | rofi |
| Display settings manager | python? | autorandr |
| Media player control | ✓ | Hautomo's playerctl |


Additional reading
------------------

- [Erase your darlings](https://grahamc.com/blog/erase-your-darlings) (powerful thought: "Leaning in to the pain")

