# gocontainer

A tiny **container runtime in ~80 lines of Go** — a learning re-implementation of
the core of `docker run`. It isolates a process with Linux **namespaces**
(UTS, PID, mount), **chroots** into a root filesystem and mounts a private
**/proc**. Standard library only, zero dependencies.

## Why this project (CV value)
- Demonstrates I understand **how containers actually work** under the hood — not
  just how to use Docker.
- Hands-on with Linux kernel primitives: `clone(2)` namespace flags, `chroot`,
  `mount`, `sethostname` — the building blocks of every container engine.
- Clean parent/child re-exec pattern; pure standard library; cross-platform build
  (a stub keeps `go build` working off-Linux, and CI builds the real thing).

## What it demonstrates
| Isolation  | Mechanism                              | Proof                                   |
|------------|----------------------------------------|-----------------------------------------|
| Hostname   | new UTS namespace + `sethostname`      | `hostname` prints `container`           |
| Filesystem | `chroot` into a rootfs                  | `/etc/os-release` shows Alpine, not host|
| Processes  | new PID namespace + private `/proc`     | the process sees itself as **PID 1**    |

## How it works
`gocontainer run <cmd>` re-executes itself (`/proc/self/exe`) as a child with
fresh namespaces (via `SysProcAttr.Cloneflags`). The child sets the hostname,
chroots into `$ROOTFS`, mounts `/proc`, then execs the command — fully isolated.

## Run it
Namespaces are a **Linux** feature, so on macOS/Windows run it inside a
privileged Linux container. With Docker:

```bash
# 1. grab a minimal root filesystem
docker export "$(docker create alpine:latest)" -o alpine-rootfs.tar

# 2. build and run gocontainer in a privileged Linux container
docker run --rm --privileged -v "$PWD":/src -w /src golang:1.26 bash -c '
  mkdir -p /rootfs && tar -xf alpine-rootfs.tar -C /rootfs
  go build -o /usr/local/bin/gocontainer .
  ROOTFS=/rootfs gocontainer run /bin/sh
'
```
On a Linux host you can build and run directly (as root):
```bash
go build -o gocontainer .
sudo ROOTFS=/path/to/rootfs ./gocontainer run /bin/sh
```

## Example
```
$ gocontainer run /bin/ps
[parent] starting [/bin/ps] (host pid 737)
[child]  isolated as pid 1, rootfs /rootfs
PID   USER     TIME  COMMAND
    1 root      0:00 /proc/self/exe child /bin/ps
    7 root      0:00 /bin/ps
```
The command sees itself as **PID 1** — a fresh PID namespace, isolated from the host.

## Next steps
- cgroups (v2) for CPU / memory / pids limits
- `pivot_root` instead of `chroot`
- user namespaces for rootless containers
- a network namespace with a veth pair
