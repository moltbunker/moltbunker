//go:build !linux

package runtime

// UserNSCompatResult reports user-namespace compatibility (defined on all
// platforms so callers and the doctor checker compile without a build-tag split).
type UserNSCompatResult struct {
	Supported bool
	Reason    string
}

// CheckUserNSCompat always reports unsupported on non-Linux platforms. This makes
// WithUserNamespace a no-op on darwin, so the spec is left unchanged there.
func CheckUserNSCompat() UserNSCompatResult {
	return UserNSCompatResult{Supported: false, Reason: "user namespaces not supported on this platform"}
}

// ResolveSubUIDRange returns a safe stub on non-Linux. It is never reached on
// darwin because WithUserNamespace only consults it when CheckUserNSCompat
// reports Supported, which it never does here.
func ResolveSubUIDRange(_ string) (hostStart uint32, size uint32, err error) {
	return 100000, 65536, nil
}
