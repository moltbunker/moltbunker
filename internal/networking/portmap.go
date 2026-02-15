package networking

import (
	"fmt"
	"sync"
)

// PortAllocator manages dynamic host port assignment from an ephemeral range.
type PortAllocator struct {
	minPort   int
	maxPort   int
	allocated map[int]bool
	mu        sync.Mutex
}

// NewPortAllocator creates a port allocator for the given range.
func NewPortAllocator(minPort, maxPort int) *PortAllocator {
	return &PortAllocator{
		minPort:   minPort,
		maxPort:   maxPort,
		allocated: make(map[int]bool),
	}
}

// Allocate finds and reserves the next available port.
func (pa *PortAllocator) Allocate() (int, error) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	for port := pa.minPort; port <= pa.maxPort; port++ {
		if !pa.allocated[port] {
			pa.allocated[port] = true
			return port, nil
		}
	}
	return 0, fmt.Errorf("no ports available in range %d-%d", pa.minPort, pa.maxPort)
}

// Release frees a previously allocated port.
func (pa *PortAllocator) Release(port int) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	delete(pa.allocated, port)
}

// IsAllocated checks if a port is currently in use.
func (pa *PortAllocator) IsAllocated(port int) bool {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	return pa.allocated[port]
}

// Count returns the number of currently allocated ports.
func (pa *PortAllocator) Count() int {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	return len(pa.allocated)
}
