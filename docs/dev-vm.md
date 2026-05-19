# Development MicroVMs

`dev-vm` runs one of five static microvm.nix slots for a project directory.
The VM identity is the slot; the project is runtime workload data.

## Architecture

```mermaid
flowchart LR
  user[Project user] -->|dev-vm ssh| ssh[ssh vsock%CID]
  root[root] -->|sudo dev-vm up/down/status/clean| launcher[dev-vm launcher]

  launcher --> state["/var/lib/dev-vms"]
  launcher --> slot["slot root export"]
  slot --> store["slot root /nix/store overlay"]
  slot --> project["slot root /home/joonas/work/project bind"]

  store --> lower["lowerdir: host /nix/store"]
  store --> upper["upperdir: slot store-upper"]

  vfio[virtiofsd] -->|single root share| vm[NixOS guest]
  vm -->|/| slot
  ssh --> vm
```

Each slot has a fixed NixOS configuration, identity, and resources:

| Slot | IP | TAP | vsock CID |
| --- | --- | --- | --- |
| `dev-vm-1` | `192.168.200.11` | `microvm-dev1` | `11` |
| `dev-vm-2` | `192.168.200.12` | `microvm-dev2` | `12` |
| `dev-vm-3` | `192.168.200.13` | `microvm-dev3` | `13` |
| `dev-vm-4` | `192.168.200.14` | `microvm-dev4` | `14` |
| `dev-vm-5` | `192.168.200.15` | `microvm-dev5` | `15` |

Slot allocation happens at launch time:

```mermaid
flowchart LR
  subgraph project["Project"]
    workdir["workdir<br/>/home/joonas/work/project"]
    up["$ dev-vm up"]
  end

  subgraph slots["MicroVM slots"]
    slot1["dev-vm-1"]
    slot2["dev-vm-2"]
    slot3["dev-vm-3<br/>example selected slot"]
    slot4["dev-vm-4"]
    slot5["dev-vm-5"]
  end

   up -- allocate free slot --> slot3
  workdir -->|bind mount| slot3
```

The host bridge is `sandbox-br` at `192.168.200.1/24`; NAT uses the host
`wired` interface.

```mermaid
flowchart LR
  subgraph vm["VM"]
    vmnic["guest NIC<br/>192.168.200.N"]
  end

  subgraph host["Host"]
    tap["TAP<br/>microvm-devN"]
    bridge["bridge<br/>sandbox-br<br/>192.168.200.1/24"]
    wired["wired NIC<br/>wired"]
  end

  subgraph lan["Physical LAN"]
    network["external network"]
  end

  vmnic --> tap
  tap --> bridge
  bridge -->|NAT| wired
  wired --> network
```

File mounts are prepared inside the slot root before `virtiofsd` starts:

```mermaid
flowchart LR
  subgraph host_inputs["Host inputs"]
    host_store["host /nix/store"]
    host_project["host project dir<br/>/home/joonas/work/project"]
  end

  subgraph slot_root["Host slot root<br/>/var/lib/dev-vms/slots/dev-vm-N/root"]
    upper["../store-upper"]
    slot_store["/nix/store<br/>overlayfs mount"]
    slot_project["/home/joonas/work/project"]
  end

  subgraph guest["Guest view"]
    guest_root["/"]
    guest_store["/nix/store<br/>writable overlay"]
    guest_project["/home/joonas/work/project<br/>writable bind"]
  end

  host_store -->|lowerdir| slot_store
  upper -->|upperdir| slot_store
  host_project -->|bind mount| slot_project

  slot_root -->|virtio-fs root share| guest_root
  guest_root --> guest_store
  guest_root --> guest_project
```

## Commands

Run lifecycle commands as root:

```sh
cd /home/joonas/work/project
sudo dev-vm up
sudo dev-vm status
sudo dev-vm down
```

Run SSH as the project user, not root:

```sh
dev-vm
dev-vm ssh
dev-vm ssh dev-vm-2
dev-vm ssh -- make test
```

Command behavior:

- `dev-vm` with no arguments starts a VM for the current directory if needed,
  then SSHes into it.
- `up [PROJECT_DIR]` allocates a free slot or reuses the slot already assigned
  to the project.
- `down [SLOT_OR_PROJECT]` stops one slot; `down --all` stops all slots.
- `status` lists slots, state, IPs, mount health, and project assignments.
- `ssh` opens an interactive login shell in the assigned project directory.
- `ssh` retries startup connection failures every 2 seconds for up to 30 seconds.
- `ssh -- REMOTE_COMMAND` runs the command without adding an implicit `cd`.
- `clean SLOT` removes per-slot writable store state; the slot must be stopped.

## Launch Flow

```mermaid
sequenceDiagram
  participant U as root shell
  participant L as dev-vm
  participant S as slot state
  participant V as virtiofsd
  participant Q as qemu

  U->>L: sudo dev-vm up PROJECT_DIR
  L->>S: lock state and select free slot
  L->>S: reset slot root
  L->>S: mount overlay at root/nix/store
  L->>S: bind project at root/PROJECT_DIR
  L->>V: start virtiofsd-run
  L->>Q: start microvm-run
  L->>S: write project, runner, pid, timestamp
```

The guest receives one virtio-fs root share from:

```text
/var/lib/dev-vms/slots/<slot>/root
```

Inside that exported root, the launcher prepares:

- `/nix/store` as a host-side overlayfs mount.
- the project directory as a writable bind mount at the same visible path.
- static additional bind mounts from the script; currently `/tmp` is
  bind-mounted to guest `/tmp`.

Project paths are normalized to the user-visible `~/work/...` path when that
path resolves to the same physical tree as `/apps/work/...`.

## State And Recovery

State lives under:

```text
/var/lib/dev-vms/
  state.lock
  slots/dev-vm-N/
    project
    runner
    qemu.pid
    virtiofsd.pid
    started-at
    root/
    run/
    store-upper/
    store-work/
```

The lock timeout is 5 seconds by default and can be changed with
`DEV_VM_LOCK_TIMEOUT`. Long-lived QEMU and virtiofsd children close the lock fd,
so `status` and `down` should work while VMs are running.

If a slot has project metadata but no running QEMU process, root commands recover
the stale slot by stopping leftover processes, unmounting slot-root mounts, and
clearing the assignment.

## Alternatives

- https://github.com/apple/container/blob/main/docs/container-machine.md
- https://github.com/herd-core/herd
- https://github.com/losfair/bake
- https://github.com/JanPokorny/locki
- [microsandbox](https://news.ycombinator.com/item?id=44135977)
- [Docker sandboxes](https://news.ycombinator.com/item?id=47042181)
