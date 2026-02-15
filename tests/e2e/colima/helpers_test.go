//go:build colima

package colima

import (
	"context"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// withProcessArgs returns an oci.SpecOpts that overrides the process command.
// This lets us run "sleep 300" or "echo hello" instead of the image default.
func withProcessArgs(args ...string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Process == nil {
			s.Process = &specs.Process{}
		}
		s.Process.Args = args
		return nil
	}
}
