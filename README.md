My personal systems' configuration and installation.

![](docs/logo.webp)

It is a collection of NixOS configurations targeting computers:

| Machine      | Platform | Desktop     |
|--------------|----------|-------------|
| Main desktop | x86      | Full        |
| Work laptop  | x86      | Full        |
| Old laptop   | x86      | Full        |
| Server       | x86      |             |
| garage       | ARM      | Thin client |
| remote USB   | ARM      |             |
| mediaplayer  | ARM      | Full        |
| Bedroom-tv   | ARM      | Thin client |
| Infra-pi     | ARM      |             |
| Security-pi  | ARM      |             |

Profiles:



Future:

- virtual desktop VMs running on server with SR-IOV capable GPU
  * I plan to have terminals from which I can connect to [VDI](https://en.wikipedia.org/wiki/Desktop_virtualization)-style desktops
  * these could be running something like [Sunshine](https://github.com/lizardbyte/sunshine) to serve the terminals over [Moonlight](https://github.com/moonlight-stream)


Usage
=====

For usage see the [justfile](justfile)


History
=======

### First iteration

A homegrown solution that:

- uses [Debootstrap](https://wiki.debian.org/Debootstrap) to bootstrap bare minimum Ubuntu installation with working [Apt](https://wiki.debian.org/Apt).
- runs ordered shell scripts to install one feature (one or multiple Debian packages)
- all of this was run in a container
- the resulting root filesystem is shipped in [OSTree](https://en.wikipedia.org/wiki/OSTree) repo for deployment to multiple computers

This was quite nice and everything was in my control, but had a few drawbacks:

- I don't like Ubuntu that much (the implementation was based on Ubuntu tooling but the ideas could be portable to other distros)
- I want to deploy the same system not just to x86 machines but ARM machines as well. While I had some basic cross-arch
  structure in place, I would have needed to pour more work into it and I'm not sure if basic Ubuntu supports newer Raspberry Pis
  and hardware acceleration correctly.
- I had issues with Wayland transition where things like Sway just didn't start
- I had some pain points in Ubuntu where some programs like Hyprland or Moonlight weren't available for the distro version I was using.
  Or the one worked with my version but the other didn't.
- I was running the exact same config in all my machines and the lack of easy structural specialization was beginning
  to be painful. Things like mounting specific-to-the-machine disks I could use `Host=...` condition in Systemd mount units
  and udev rules were just fine even when they were on machines where hardware that the rule targeted didn't exist..
  but it was certainly beginning to be gross.


### The next solution

This was to solve the previous pain points.

Currently I am investigating NixOS for this. When I started the first solution I knew NixOS existed and had significant overlap.
But the thing is that I really hate Nix still to this day - for multiple reasons.
But even though I hate it, it is possible that it is the least worst tool out there. 🤷

