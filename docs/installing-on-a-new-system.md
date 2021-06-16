Installing on a new system
==========================

Steps for installing joonas-sys on a new system.

It is assumed that you have a USB stick that you want to install the readonly system on, and use the
computer's full disk to host the `/persist`. I found out a
[small form-factor USB flash memory](https://www.samsung.com/us/computing/memory-storage/usb-flash-drives/usb-3-1-flash-drive-fit-plus-256gb-muf-256ab-am/)
to be nice.

Mount the USB stick on a different computer.


Partitioning
------------

Partition the USB stick.

512 MB for the ESP. The remaining will be 50-50 % split to active/passive partitions.

```
$ gdisk /dev/DISK
> o (for new partition table)

> n (for new partition)
> last sector: +512M
> hex code: EF00
```

Now for the two data partitions. The next partition you create, it suggests the last sector to fill
the rest of the drive, like this:

> Last sector (1050624-125313249, default = 125313249) or {+-}size{KMGTP}:

If you subtract 125313249-1050624, you get 124262625 as the suggested partition size. Let's halve that:

124262625/2=62131312. Therefore enter 1050624 + 62131312 as the last sector.

You'll now have something like this:

```
Number  Start (sector)    End (sector)  Size       Code  Name
   1            2048         1050623   512.0 MiB   EF00  EFI system partition
   2         1050624        62131312   29.1 GiB    8300  Linux filesystem
   3        62132224       125313249   30.1 GiB    8300  Linux filesystem
```

Make filesystems
-----------------

```
$ mkfs.fat -F32 -n ESP-USB-FIT /dev/sdc1
$ mkfs.ext4 -L system_a /dev/sdc2
$ mkfs.ext4 -L system_b /dev/sdc3
```

The system-to-be-provisioned is known as system label `provision` in system tooling.
Make sure system registry points to correct devices!

Now create ESP template: `$ jsys esp-create-template provision`

Then flash the system: `$ jsys flash provision`


Create persist partition on the new computer
--------------------------------------------

Now plug the USB stick to the new computer. You'll be greeted by the bootloader.

If you tried to boot into `system_a`, the boot process would panic because there isn't a
`/persist` partition yet. We need to create it.

Boot into `system_a`'s initramfs to create the partition.
You have `$ lsblk` and `$ mkfs.ext4` at your disposal.

Format persist partition: `$ mkfs.ext4 -L persist /dev/SOMETHING`

Mount the persist partition at `/persist`

Write a hostname to `/persist/apps/SYSTEM_nobackup/hostname` 

Docker doesn't start without this:

```
$ mkdir -p /persist/apps/docker/data_nobackup
```

A quick `$ poweroff` later and you should be able to boot into the new computer.


Persist partition finishing touches
-----------------------------------

Set wallpaper at `/persist/apps/SYSTEM_nobackup/background.png`

Symlink `/persist/apps/SYSTEM_nobackup/cpu_temp` to `/sys/class/hwmon/hwmon<NUM>temp<NUM>_input`
that represents your CPU temp.
