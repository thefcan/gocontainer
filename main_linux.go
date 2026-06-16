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
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gocontainer run <cmd> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}

// run is the parent process: it re-executes this binary as "child" inside fresh
// namespaces, so the child starts already isolated.
func run() {
	fmt.Printf("[parent] starting %v (host pid %d)\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Fresh UTS (hostname), PID and mount namespaces for the child.
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

// child runs inside the new namespaces: it sets the hostname, chroots into the
// root filesystem, mounts /proc and finally execs the requested command.
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

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	runErr := cmd.Run()

	_ = syscall.Unmount("/proc", 0) // best-effort cleanup
	must(runErr)
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "gocontainer: %v\n", err)
		os.Exit(1)
	}
}
