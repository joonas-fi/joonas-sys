Recovery boot
=============

Let's say that something is borked on your ESP and either the kernel or the initrd is broken.


How to boot "manually"
----------------------

Our ESP contains u-root, which you can enter from eEFInd.

Choose emergency recovery from eEFInd

You're now in u-root. List block devices:

ls /dev | grep sd

mount /dev/sda /system

cd /system/boot

kexec --load --initrd initrd.img --cmdline "root=LABEL=system_a ro" vmlinuz
kexec --exec

(--cmdline actually being the one that works in eEFInd)