package kubernetes

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"
)

// PortAllocator hands out free TCP ports within a configured range. It tracks
// in-process reservations so a single allocator instance never returns the
// same port twice, even before the caller has bound it.
type PortAllocator struct {
	mu       sync.Mutex
	start    int
	end      int
	cursor   int
	reserved map[int]bool
}

// NewPortAllocator returns an allocator that scans [start, end] inclusive.
// If the range is invalid it falls back to 35000-60000.
func NewPortAllocator(start, end int) *PortAllocator {
	if start <= 0 || end <= 0 || end < start {
		start, end = 35000, 60000
	}
	return &PortAllocator{
		start:    start,
		end:      end,
		cursor:   start,
		reserved: map[int]bool{},
	}
}

// Allocate returns the first port in [start, end] that the allocator has not
// already handed out and that the OS reports as bindable.
func (pa *PortAllocator) Allocate() (int, error) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	total := pa.end - pa.start + 1
	for tried := 0; tried < total; tried++ {
		port := pa.cursor
		pa.cursor++
		if pa.cursor > pa.end {
			pa.cursor = pa.start
		}
		if pa.reserved[port] {
			continue
		}
		if !portFree(port) {
			continue
		}
		pa.reserved[port] = true
		return port, nil
	}
	return 0, fmt.Errorf("no free port available in range %d-%d", pa.start, pa.end)
}

// Release marks a port as available again. Callers should release a port once
// the process holding it has exited.
func (pa *PortAllocator) Release(port int) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	delete(pa.reserved, port)
}

func portFree(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

// PortForward wraps a `kubectl port-forward` child process targeting a single
// pod. Local port selection is delegated to a PortAllocator; the same port is
// released on Stop.
type PortForward struct {
	namespace  string
	podName    string
	localPort  int
	remotePort int
	allocator  *PortAllocator

	mu      sync.Mutex
	running bool
	cmd     *exec.Cmd
}

// NewPortForward builds a forwarder; call Start to actually launch kubectl.
func NewPortForward(allocator *PortAllocator, namespace, podName string, remotePort int) *PortForward {
	return &PortForward{
		namespace:  namespace,
		podName:    podName,
		remotePort: remotePort,
		allocator:  allocator,
	}
}

// Start picks a free local port (retrying up to maxAttempts times if kubectl
// fails to bind it) and launches `kubectl port-forward` in the background.
func (pf *PortForward) Start() error {
	pf.mu.Lock()
	if pf.running {
		pf.mu.Unlock()
		return fmt.Errorf("port forward already running")
	}
	pf.mu.Unlock()

	const maxAttempts = 5
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		port, err := pf.allocator.Allocate()
		if err != nil {
			return fmt.Errorf("port allocation failed: %w", err)
		}

		args := []string{
			"port-forward",
			"-n", pf.namespace,
			fmt.Sprintf("pod/%s", pf.podName),
			fmt.Sprintf("%d:%d", port, pf.remotePort),
		}
		cmd := exec.Command("kubectl", args...)

		log.Printf("[PortForward] attempt %d: kubectl %v", attempt, args)
		if err := cmd.Start(); err != nil {
			pf.allocator.Release(port)
			lastErr = fmt.Errorf("kubectl start failed: %w", err)
			continue
		}

		// Give kubectl a moment to bind. If it dies immediately (port still in
		// use, pod gone, etc.) retry with another port.
		time.Sleep(1500 * time.Millisecond)
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			pf.allocator.Release(port)
			lastErr = fmt.Errorf("kubectl exited immediately on port %d", port)
			continue
		}

		pf.mu.Lock()
		pf.cmd = cmd
		pf.localPort = port
		pf.running = true
		pf.mu.Unlock()
		log.Printf("[PortForward] established 127.0.0.1:%d -> %s/%s:%d", port, pf.namespace, pf.podName, pf.remotePort)
		return nil
	}
	return fmt.Errorf("port-forward failed after %d attempts: %v", maxAttempts, lastErr)
}

// Stop terminates the kubectl child and releases the allocated port.
func (pf *PortForward) Stop() {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	if !pf.running {
		return
	}
	if pf.cmd != nil && pf.cmd.Process != nil {
		_ = pf.cmd.Process.Kill()
		_, _ = pf.cmd.Process.Wait()
	}
	if pf.localPort > 0 {
		pf.allocator.Release(pf.localPort)
	}
	pf.running = false
	log.Printf("[PortForward] stopped 127.0.0.1:%d -> %s/%s", pf.localPort, pf.namespace, pf.podName)
}

// GetLocalPort returns the port chosen by Start (0 before Start succeeds).
func (pf *PortForward) GetLocalPort() int {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	return pf.localPort
}

// GetLocalAddress returns the dial-able 127.0.0.1:<port> for this forward.
func (pf *PortForward) GetLocalAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", pf.GetLocalPort())
}

// IsRunning reports whether the underlying kubectl process is still alive.
func (pf *PortForward) IsRunning() bool {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	return pf.running
}
