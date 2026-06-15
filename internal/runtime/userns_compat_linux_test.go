//go:build linux

package runtime

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

// writeTempFile writes content to a fresh temp file and returns its path.
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func fixedUser(name string) func() (*user.User, error) {
	return func() (*user.User, error) { return &user.User{Username: name}, nil }
}

func TestCheckUserNSCompat(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	tests := []struct {
		name      string
		paths     usernsPaths
		wantOK    bool
		reasonHas string
	}{
		{
			name: "clone disabled",
			paths: usernsPaths{
				unprivilegedCloneFile: writeTempFile(t, "clone", "0\n"),
				maxUserNamespacesFile: missing,
				subUIDFile:            writeTempFile(t, "subuid", "alice:100000:65536\n"),
				subGIDFile:            writeTempFile(t, "subgid", "alice:100000:65536\n"),
				currentUser:           fixedUser("alice"),
			},
			wantOK:    false,
			reasonHas: "unprivileged_userns_clone disabled",
		},
		{
			name: "max_user_namespaces zero",
			paths: usernsPaths{
				unprivilegedCloneFile: missing,
				maxUserNamespacesFile: writeTempFile(t, "maxuserns", "0\n"),
				subUIDFile:            writeTempFile(t, "subuid", "alice:100000:65536\n"),
				subGIDFile:            writeTempFile(t, "subgid", "alice:100000:65536\n"),
				currentUser:           fixedUser("alice"),
			},
			wantOK:    false,
			reasonHas: "max_user_namespaces=0",
		},
		{
			name: "missing subuid",
			paths: usernsPaths{
				unprivilegedCloneFile: missing,
				maxUserNamespacesFile: missing,
				subUIDFile:            missing,
				subGIDFile:            writeTempFile(t, "subgid", "alice:100000:65536\n"),
				currentUser:           fixedUser("alice"),
			},
			wantOK:    false,
			reasonHas: "no subuid range",
		},
		{
			name: "subuid range too small",
			paths: usernsPaths{
				unprivilegedCloneFile: missing,
				maxUserNamespacesFile: missing,
				subUIDFile:            writeTempFile(t, "subuid", "alice:100000:1024\n"),
				subGIDFile:            writeTempFile(t, "subgid", "alice:100000:65536\n"),
				currentUser:           fixedUser("alice"),
			},
			wantOK:    false,
			reasonHas: "no subuid range",
		},
		{
			name: "missing subgid",
			paths: usernsPaths{
				unprivilegedCloneFile: missing,
				maxUserNamespacesFile: missing,
				subUIDFile:            writeTempFile(t, "subuid", "alice:100000:65536\n"),
				subGIDFile:            missing,
				currentUser:           fixedUser("alice"),
			},
			wantOK:    false,
			reasonHas: "no subgid range",
		},
		{
			name: "healthy",
			paths: usernsPaths{
				unprivilegedCloneFile: writeTempFile(t, "clone", "1\n"),
				maxUserNamespacesFile: writeTempFile(t, "maxuserns", "15000\n"),
				subUIDFile:            writeTempFile(t, "subuid", "alice:100000:65536\n"),
				subGIDFile:            writeTempFile(t, "subgid", "alice:100000:65536\n"),
				currentUser:           fixedUser("alice"),
			},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := checkUserNSCompat(tt.paths)
			if res.Supported != tt.wantOK {
				t.Fatalf("Supported = %v, want %v (reason: %q)", res.Supported, tt.wantOK, res.Reason)
			}
			if tt.reasonHas != "" && !contains(res.Reason, tt.reasonHas) {
				t.Errorf("Reason = %q, want substring %q", res.Reason, tt.reasonHas)
			}
		})
	}
}

func TestResolveSubIDRange_ParsesMultiEntry(t *testing.T) {
	path := writeTempFile(t, "subuid", strings.Join([]string{
		"# a comment",
		"bob:200000:65536",
		"alice:100000:1024",   // too small, skipped
		"alice:300000:131072", // first usable range for alice
	}, "\n")+"\n")

	start, size, err := resolveSubIDRange(path, "alice")
	if err != nil {
		t.Fatalf("resolveSubIDRange error: %v", err)
	}
	if start != 300000 || size != 131072 {
		t.Errorf("got (start=%d size=%d), want (300000, 131072)", start, size)
	}

	if _, _, err := resolveSubIDRange(path, "carol"); err == nil {
		t.Error("expected error for a user with no range")
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
