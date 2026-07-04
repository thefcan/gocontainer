package main

import "testing"

func TestResolveRootfs(t *testing.T) {
	if got := resolveRootfs(""); got != defaultRootfs {
		t.Errorf("resolveRootfs(%q) = %q, want %q", "", got, defaultRootfs)
	}
	if got := resolveRootfs("/custom/root"); got != "/custom/root" {
		t.Errorf("resolveRootfs(%q) = %q, want it unchanged", "/custom/root", got)
	}
}

func TestResolvePath(t *testing.T) {
	if got := resolvePath(""); got != defaultPath {
		t.Errorf("resolvePath(%q) = %q, want the default PATH", "", got)
	}
	custom := "/opt/bin:/usr/bin"
	if got := resolvePath(custom); got != custom {
		t.Errorf("resolvePath(%q) = %q, want it unchanged", custom, got)
	}
}

func TestDefaultsAreNonEmpty(t *testing.T) {
	if defaultRootfs == "" || containerHostname == "" || defaultPath == "" {
		t.Fatal("default configuration constants must not be empty")
	}
}
