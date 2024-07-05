MBR (points to /boot)


https://current.workingdirectory.net/posts/2009/grub-on-usb/


p
Disk /dev/loop0: 2048 MB, 2147483648 bytes, 4194304 sectors
261 cylinders, 255 heads, 63 sectors/track
Units: sectors of 1 * 512 = 512 bytes

Device     Boot StartCHS    EndCHS        StartLBA     EndLBA    Sectors  Size Id Type
/dev/loop0p1 *  0,1,1       261,21,16           63    4194303    4194241 2047M  b Win95 FAT32


sudo mkfs.fat -F 32 /dev/loop0p1 -n LEGACYBOOT

sudo mount /dev/loop0p1 /tmp/tikku

sudo grub-install --target=i386-pc --root-directory=/tmp/tikku /dev/loop0

mkdir /tmp/tikku/boot

sudo qemu-system-x86_64 -drive format=raw,file=/dev/loop0
