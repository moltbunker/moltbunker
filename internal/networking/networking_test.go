package networking

import (
	"sync"
	"testing"
)

// --- PortAllocator tests ---

func TestPortAllocator_Allocate(t *testing.T) {
	pa := NewPortAllocator(50000, 50002)

	p1, err := pa.Allocate()
	if err != nil {
		t.Fatalf("first allocate: %v", err)
	}
	if p1 != 50000 {
		t.Errorf("first port = %d, want 50000", p1)
	}

	p2, err := pa.Allocate()
	if err != nil {
		t.Fatalf("second allocate: %v", err)
	}
	if p2 != 50001 {
		t.Errorf("second port = %d, want 50001", p2)
	}

	p3, err := pa.Allocate()
	if err != nil {
		t.Fatalf("third allocate: %v", err)
	}
	if p3 != 50002 {
		t.Errorf("third port = %d, want 50002", p3)
	}

	// Should be exhausted
	_, err = pa.Allocate()
	if err == nil {
		t.Fatal("should fail when all ports exhausted")
	}
}

func TestPortAllocator_Release(t *testing.T) {
	pa := NewPortAllocator(50000, 50000) // Only 1 port

	p, _ := pa.Allocate()
	if _, err := pa.Allocate(); err == nil {
		t.Fatal("should fail when exhausted")
	}

	pa.Release(p)

	p2, err := pa.Allocate()
	if err != nil {
		t.Fatalf("allocate after release: %v", err)
	}
	if p2 != p {
		t.Errorf("re-allocated port = %d, want %d", p2, p)
	}
}

func TestPortAllocator_IsAllocated(t *testing.T) {
	pa := NewPortAllocator(50000, 50010)

	if pa.IsAllocated(50000) {
		t.Error("50000 should not be allocated yet")
	}

	pa.Allocate()
	if !pa.IsAllocated(50000) {
		t.Error("50000 should be allocated")
	}
}

func TestPortAllocator_Count(t *testing.T) {
	pa := NewPortAllocator(50000, 50010)

	if pa.Count() != 0 {
		t.Errorf("count = %d, want 0", pa.Count())
	}

	pa.Allocate()
	pa.Allocate()
	if pa.Count() != 2 {
		t.Errorf("count = %d, want 2", pa.Count())
	}

	pa.Release(50000)
	if pa.Count() != 1 {
		t.Errorf("count after release = %d, want 1", pa.Count())
	}
}

func TestPortAllocator_Concurrent(t *testing.T) {
	pa := NewPortAllocator(50000, 50099) // 100 ports

	var wg sync.WaitGroup
	allocated := make(chan int, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := pa.Allocate()
			if err != nil {
				return
			}
			allocated <- p
		}()
	}
	wg.Wait()
	close(allocated)

	seen := make(map[int]bool)
	for p := range allocated {
		if seen[p] {
			t.Errorf("duplicate port allocated: %d", p)
		}
		seen[p] = true
	}
	if len(seen) != 100 {
		t.Errorf("allocated %d unique ports, want 100", len(seen))
	}
}

// --- NetworkManager tests ---

func TestNetworkManager_SetupTeardown(t *testing.T) {
	nm := NewNetworkManager()

	ports := []ExposedPort{
		{ContainerPort: 80},
		{ContainerPort: 443},
	}

	net, err := nm.SetupNetwork("deploy-1", ports)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if net == nil {
		t.Fatal("network should not be nil")
	}

	// Should be retrievable
	got, ok := nm.GetNetwork("deploy-1")
	if !ok || got == nil {
		t.Fatal("network should be retrievable after setup")
	}

	// Duplicate should fail
	_, err = nm.SetupNetwork("deploy-1", ports)
	if err == nil {
		t.Fatal("duplicate setup should fail")
	}

	// Teardown
	if err := nm.TeardownNetwork("deploy-1"); err != nil {
		t.Fatalf("teardown: %v", err)
	}

	// Should be gone
	_, ok = nm.GetNetwork("deploy-1")
	if ok {
		t.Fatal("network should be gone after teardown")
	}
}

func TestNetworkManager_TeardownNonexistent(t *testing.T) {
	nm := NewNetworkManager()
	// Should not error
	if err := nm.TeardownNetwork("nonexistent"); err != nil {
		t.Fatalf("teardown nonexistent: %v", err)
	}
}

func TestNetworkManager_PortAllocation(t *testing.T) {
	nm := NewNetworkManager()

	ports := []ExposedPort{
		{ContainerPort: 8080},                  // auto-assign
		{ContainerPort: 9090, HostPort: 30000}, // fixed
	}

	net, err := nm.SetupNetwork("deploy-2", ports)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Container port 8080 should be mapped to an ephemeral port
	hp, ok := net.ResolvePort(8080)
	if !ok {
		t.Fatal("port 8080 should be mapped")
	}
	if hp < 49152 || hp > 65535 {
		t.Errorf("auto-assigned port %d not in ephemeral range", hp)
	}

	// Container port 9090 should map to fixed 30000
	hp, ok = net.ResolvePort(9090)
	if !ok {
		t.Fatal("port 9090 should be mapped")
	}
	if hp != 30000 {
		t.Errorf("fixed port = %d, want 30000", hp)
	}

	// Unknown port
	_, ok = net.ResolvePort(1234)
	if ok {
		t.Error("port 1234 should not be mapped")
	}
}

// --- Fallback network tests (non-Linux) ---

func TestFallbackNetwork_ContainerIP(t *testing.T) {
	net := newContainerNetwork("test", []ExposedPort{
		{ContainerPort: 80, HostPort: 8080},
	})
	if net.ContainerIP() != "127.0.0.1" {
		t.Errorf("container IP = %q, want 127.0.0.1", net.ContainerIP())
	}
}

func TestFallbackNetwork_SetupTeardown(t *testing.T) {
	net := newContainerNetwork("test", nil)
	if err := net.Setup(); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := net.Teardown(); err != nil {
		t.Fatalf("teardown: %v", err)
	}
}

// --- Port exhaustion rollback test ---

func TestNetworkManager_PortExhaustionRollback(t *testing.T) {
	// Create a manager with a very small port range
	nm := &NetworkManager{
		networks:  make(map[string]ContainerNetwork),
		portAlloc: NewPortAllocator(60000, 60000), // only 1 port available
	}

	// First setup uses the only port
	_, err := nm.SetupNetwork("dep-1", []ExposedPort{{ContainerPort: 80}})
	if err != nil {
		t.Fatalf("first setup: %v", err)
	}

	// Second setup should fail (port exhausted) and NOT leave partial state
	_, err = nm.SetupNetwork("dep-2", []ExposedPort{{ContainerPort: 80}})
	if err == nil {
		t.Fatal("second setup should fail with port exhaustion")
	}

	// dep-2 should not be in the network map
	if _, ok := nm.GetNetwork("dep-2"); ok {
		t.Error("failed setup should not leave network in map")
	}
}
