//go:build linux

// Command gocontainer is a tiny container runtime. It isolates a process using
// Linux namespaces (UTS, PID, mount), chroots into a root filesystem and mounts
// a private /proc — a learning re-implementation of the core of `docker run`.
//
// Usage:
//
//	gocontainer run <cmd> [args...]
//
// The root filesystem is taken from the ROOTFS env var (default /rootfs).
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			usage()
		}
		run()
	case "child":
		if len(os.Args) < 3 {
			usage()
		}
		child()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
	}
}

// run is the parent process: it re-executes this binary as "child" inside fresh
// namespaces, so the child starts already isolated. It forwards the child's
// exit code, so `gocontainer run` behaves like the command it wraps.
func run() {
	fmt.Printf("[parent] starting %v (host pid %d)\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Fresh UTS (hostname), PID and mount namespaces for the child.
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.ExitCode())
		}
		fatal(err)
	}
}

// child runs inside the new namespaces: it sets the hostname, chroots into the
// root filesystem, mounts /proc, then replaces itself with the requested command
// via exec — so the command runs as PID 1, exactly like a real container's init.
func child() {
	rootfs := os.Getenv("ROOTFS")
	if rootfs == "" {
		rootfs = "/rootfs"
	}
	fmt.Printf("[child]  isolated as pid %d, rootfs %s\n", os.Getpid(), rootfs)

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot(rootfs))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))

	// Give bare command names (e.g. "sh") a sane PATH to resolve against inside
	// the new root filesystem.
	if os.Getenv("PATH") == "" {
		os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	}

	path, err := exec.LookPath(os.Args[2])
	must(err)

	// Replace this process image with the command. It inherits our PID (1), so
	// the workload itself becomes the container's init process. The private
	// /proc mount is torn down automatically when the namespace exits.
	must(syscall.Exec(path, os.Args[2:], os.Environ()))
}

func must(err error) {
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "gocontainer: %v\n", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gocontainer run <cmd> [args...]")
	os.Exit(1)
}
