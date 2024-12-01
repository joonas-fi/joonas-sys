Memory safety roadmap
=====================

Why?
----

[What is memory safety and why does it matter?](https://www.memorysafety.org/docs/memory-safety/)


Status
------

Here's the most essential components of the stack:

| Component | Memory safe | Program | Notes |
|-----------|-------------|---------|-------|
| EFI bootloader | | rEFInd | Investigate [u-root/systemboot](https://github.com/u-root/u-root#systemboot) |
| Kernel | | Linux kernel | |
| Early userspace | | Ubuntu initramfs | [Garbage](https://twitter.com/joonas_fi/status/1368276201643577347) |
| Init system | | systemd | |
| Display server | | Xorg | [Waiting for Wayland to mature](https://twitter.com/dave_universetf/status/1357825910674657282) |
| Display manager | | LightDM | [Info about display mangers](https://wiki.archlinux.org/index.php/display_manager) |
| Greeter | | [slick-greeter](https://github.com/linuxmint/slick-greeter) | |
| Window manager |  | i3 | [Investigate memory safe alternatives](https://users.rust-lang.org/t/is-there-a-tiling-window-manager-for-linux-that-is-written-and-configurable-in-rust/4407) |
| Compositor |  | compton | |
| Screensaver |  | xfce4-screensaver | |
| Screenshot app |  | Flameshot | |
| Screen recorder |  | Peek | |
| Notification daemon |  | xfce4-notifyd | |
| Clipboard manager | ✓ | Clipcat | |
| Terminal | ✓ | Alacritty | |
| Program launcher | | rofi | |
| Display settings manager | ✓ | autorandr | |
| Media player control | ✓ | Hautomo's playerctl | |
| VPN / mesh connectivity | ✓ | Tailscale | |
