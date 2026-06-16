//go:build !linux

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "gocontainer only runs on Linux (it uses Linux namespaces, chroot and /proc).")
	fmt.Fprintln(os.Stderr, "Run it inside a privileged Linux container — see the Docker command in the README.")
	os.Exit(1)
}
