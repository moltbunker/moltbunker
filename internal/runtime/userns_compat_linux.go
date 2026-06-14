//go:build linux

package runtime

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

// minSubIDRange is the smallest subordinate UID/GID range that yields a usable
// full container UID space (0..65535). Anything smaller cannot host a standard
// rootful-in-userns container.
const minSubIDRange = 65536

// UserNSCompatResult reports whether unprivileged user namespaces can be used for
// tenant workloads on this host, and if not, a human-readable reason.
type UserNSCompatResult struct {
	Supported bool
	Reason    string
}

// usernsPaths bundles the kernel/distro files CheckUserNSCompat inspects. The
// defaults point at the real procfs/config; tests inject temp-file substitutes.
type usernsPaths struct {
	unprivilegedCloneFile string // /proc/sys/kernel/unprivileged_userns_clone (Debian/Ubuntu knob)
	maxUserNamespacesFile string // /proc/sys/user/max_user_namespaces
	subUIDFile            string // /etc/subuid
	subGIDFile            string // /etc/subgid
	currentUser           func() (*user.User, error)
}

func defaultUsernsPaths() usernsPaths {
	return usernsPaths{
		unprivilegedCloneFile: "/proc/sys/kernel/unprivileged_userns_clone",
		maxUserNamespacesFile: "/proc/sys/user/max_user_namespaces",
		subUIDFile:            "/etc/subuid",
		subGIDFile:            "/etc/subgid",
		currentUser:           user.Current,
	}
}

// CheckUserNSCompat reports whether unprivileged user namespaces are usable for
// tenant deployment workloads on this host. It is a compat guard: a {false, reason}
// result means WithUserNamespace degrades to a no-op (container runs without a
// userns) instead of failing the deploy on a distro that disables the feature.
func CheckUserNSCompat() UserNSCompatResult {
	return checkUserNSCompat(defaultUsernsPaths())
}

func checkUserNSCompat(p usernsPaths) UserNSCompatResult {
	// (1) Debian/Ubuntu gate: kernel.unprivileged_userns_clone == 0 disables it.
	if v, ok := readSysctlInt(p.unprivilegedCloneFile); ok && v == 0 {
		return UserNSCompatResult{Supported: false, Reason: "unprivileged_userns_clone disabled (sysctl kernel.unprivileged_userns_clone=0)"}
	}

	// (2) Generic gate: user.max_user_namespaces == 0 disables it everywhere.
	if v, ok := readSysctlInt(p.maxUserNamespacesFile); ok && v == 0 {
		return UserNSCompatResult{Supported: false, Reason: "max_user_namespaces=0 (sysctl user.max_user_namespaces=0)"}
	}

	uname := "root"
	if p.currentUser != nil {
		if u, err := p.currentUser(); err == nil && u.Username != "" {
			uname = u.Username
		}
	}

	// (3) A subordinate UID range >= 65536 must exist for the daemon's user.
	if _, _, err := resolveSubIDRange(p.subUIDFile, uname); err != nil {
		return UserNSCompatResult{Supported: false, Reason: fmt.Sprintf("no subuid range >=%d for user %q in %s", minSubIDRange, uname, p.subUIDFile)}
	}
	// (4) ...and a matching subordinate GID range.
	if _, _, err := resolveSubIDRange(p.subGIDFile, uname); err != nil {
		return UserNSCompatResult{Supported: false, Reason: fmt.Sprintf("no subgid range >=%d for user %q in %s", minSubIDRange, uname, p.subGIDFile)}
	}

	return UserNSCompatResult{Supported: true}
}

// ResolveSubUIDRange returns the first /etc/subuid range >= 65536 for the given
// user: (hostStart, size). Used by WithUserNamespace to pick the host UID base
// for the container's 0..65535 mapping.
func ResolveSubUIDRange(username string) (hostStart uint32, size uint32, err error) {
	return resolveSubIDRange("/etc/subuid", username)
}

// resolveSubIDRange parses a subuid/subgid file for the first range >= minSubIDRange
// belonging to username (matched by name or numeric UID/GID). Lines look like:
//
//	username:100000:65536
func resolveSubIDRange(path, username string) (uint32, uint32, error) {
	// #nosec G304 -- path is a fixed config constant (/etc/subuid|subgid) or a test temp file, never request input.
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 3 {
			continue
		}
		if parts[0] != username {
			continue
		}
		start, serr := strconv.ParseUint(parts[1], 10, 32)
		if serr != nil {
			continue
		}
		size, zerr := strconv.ParseUint(parts[2], 10, 32)
		if zerr != nil {
			continue
		}
		if size < minSubIDRange {
			continue
		}
		return uint32(start), uint32(size), nil
	}
	return 0, 0, fmt.Errorf("no subordinate ID range >=%d for %q in %s", minSubIDRange, username, path)
}

// readSysctlInt reads a single-integer sysctl-style file. ok is false when the
// file is absent or unparseable (callers treat that as "not constrained").
func readSysctlInt(path string) (val int, ok bool) {
	// #nosec G304 -- path is a fixed /proc/sys constant or a test temp file, never request input.
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, false
	}
	return n, true
}
