package kubernetes

import (
	"bufio"
	"fmt"
	"io"
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
	contextName string
	namespace   string
	podName     string
	localPort   int
	remotePort  int
	allocator   *PortAllocator

	mu      sync.Mutex
	running bool
	cmd     *exec.Cmd
}

// NewPortForward builds a forwarder; call Start to actually launch kubectl.
// contextName is passed via --context so kubectl uses the same cluster the
// caller discovered the pod from, regardless of the kubeconfig's
// current-context. Empty contextName falls back to the kubeconfig default.
func NewPortForward(allocator *PortAllocator, contextName, namespace, podName string, remotePort int) *PortForward {
	return &PortForward{
		contextName: contextName,
		namespace:   namespace,
		podName:     podName,
		remotePort:  remotePort,
		allocator:   allocator,
	}
}

// readyTimeout is how long we wait for kubectl's tunnel to start accepting
// TCP connections before giving up on a single attempt. AKS-style clusters
// can take several seconds to set up the SPDY stream after writing the
// "Forwarding from..." line; a short fixed sleep is not enough.
const readyTimeout = 10 * time.Second

// Start picks a free local port (retrying up to maxAttempts times if kubectl
// fails to bind it) and launches `kubectl port-forward` in the background.
// kubectl's stdout/stderr are streamed to our log so binding failures are
// visible. We declare success only after a TCP dial to the local port
// succeeds — kubectl prints "Forwarding from..." optimistically, but the
// actual tunnel may not be ready for a few hundred ms (longer on remote
// clusters), so trusting that line alone leads to "actively refused" errors
// when grpcurl races us.
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

		// --address 127.0.0.1 forces IPv4 binding. Without it kubectl may bind
		// only to [::1] on some Windows configurations, which grpcurl (dialing
		// 127.0.0.1) cannot reach.
		args := []string{"port-forward", "--address", "127.0.0.1"}
		if pf.contextName != "" {
			args = append(args, "--context", pf.contextName)
		}
		args = append(args,
			"-n", pf.namespace,
			fmt.Sprintf("pod/%s", pf.podName),
			fmt.Sprintf("%d:%d", port, pf.remotePort),
		)
		cmd := exec.Command("kubectl", args...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			pf.allocator.Release(port)
			return fmt.Errorf("stdout pipe: %w", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			pf.allocator.Release(port)
			return fmt.Errorf("stderr pipe: %w", err)
		}

		log.Printf("[PortForward] attempt %d: kubectl %v", attempt, args)
		if err := cmd.Start(); err != nil {
			pf.allocator.Release(port)
			lastErr = fmt.Errorf("kubectl start failed: %w", err)
			continue
		}

		go streamLines(stdout, fmt.Sprintf("[PortForward:%d/stdout]", port))
		go streamLines(stderr, fmt.Sprintf("[PortForward:%d/stderr]", port))

		if err := waitForListening(port, readyTimeout); err != nil {
			log.Printf("[PortForward] port %d never came up: %v; killing kubectl and retrying", port, err)
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
			pf.allocator.Release(port)
			lastErr = err
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

// streamLines copies a kubectl pipe to our log, line-by-line. Returns when
// the pipe hits EOF (kubectl exited or was killed).
func streamLines(r io.Reader, prefix string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("%s %s", prefix, scanner.Text())
	}
}

// waitForListening polls a TCP dial against 127.0.0.1:port until it succeeds
// or the timeout elapses. Used to confirm kubectl actually bound the port,
// not just printed its readiness line.
func waitForListening(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("port not listening after %s: %v", timeout, lastErr)
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
