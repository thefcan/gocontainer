package main

// Default configuration for the container's isolated environment. These mirror
// what the runtime provides when the caller does not override them through the
// environment. They live in a build-tag-free file so the logic stays testable
// on any platform, even though the runtime itself only executes on Linux.
const (
	// defaultRootfs is the root filesystem the child chroots into when ROOTFS
	// is unset.
	defaultRootfs = "/rootfs"

	// containerHostname is the hostname set inside the new UTS namespace.
	containerHostname = "container"

	// defaultPath is a sane PATH for resolving bare command names (e.g. "sh")
	// inside the container when the caller passes none through.
	defaultPath = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
)

// resolveRootfs returns the root filesystem to chroot into: the caller-provided
// value (typically the ROOTFS env var) when non-empty, otherwise defaultRootfs.
func resolveRootfs(envRootfs string) string {
	if envRootfs == "" {
		return defaultRootfs
	}
	return envRootfs
}

// resolvePath returns the PATH to expose inside the container: the caller's own
// PATH when set, otherwise defaultPath so bare command names still resolve.
func resolvePath(envPath string) string {
	if envPath == "" {
		return defaultPath
	}
	return envPath
}
