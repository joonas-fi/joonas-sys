esp-template-disk-512M.gz.img
-----------------------------

Is a disk image with GPT partitioning. It has one 512 MB partition - an empty ESP partition.

Was made with commands:

```console
$ truncate -s 512M esp.img

$ gdisk esp.img

> o
> n
> Hex code or GUID: EF00
> w

$ losetup --find --partscan esp.img

$ mkfs.fat -F32 -n ESP /dev/loop1p1

$ cat /dev/loop1 | gzip > esp-template-disk-512M.gz.img
```
