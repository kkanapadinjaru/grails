package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"path"
	"strings"
	"sync"
	"time"

	"grails/auth"
	"grails/config"
	grpcrefl "grails/grpc"
	"grails/kubernetes"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// hostContextID is the magic context name we use for the synthetic localhost
// entry surfaced in the ENVIRONMENT dropdown when something is listening on
// the host gRPC port. It must not collide with any real kubeconfig context
// name.
const hostContextID = "__localhost__"

// hostDisplayName is what the user sees in the ENVIRONMENT dropdown. Kept
// lowercase to match the casing of kubeconfig contexts like "docker-desktop".
const hostDisplayName = "localhost"

// hostNamespace is the single "namespace" we surface when connected to the
// host. It's just a label — there's no real namespace.
const hostNamespace = "localhost"

// hostProbePort is the single port we probe on 127.0.0.1 to decide whether
// to surface the localhost environment. Hardcoded (rather than reusing
// cfg.GrpcPorts) because k8s services and host services have different
// conventions — k8s pods often expose 5001/5002, while host processes
// conventionally use just 5001.
const hostProbePort = 5001

// ClusterInfo is the cluster summary exposed to the frontend.
type ClusterInfo struct {
	Name    string `json:"name"`
	Context string `json:"context"`
	Server  string `json:"server"`
}

// NamespaceInfo describes one configured namespace and whether the user can
// list services within it.
type NamespaceInfo struct {
	Name    string `json:"name"`
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// GrpcServiceInfo is the per-service entry rendered in the SERVICE dropdown.
// DisplayName is "<app.kubernetes.io/name>:<host_port>"; ServiceName is the
// canonical gRPC service (e.g. "grpc.health.v1.Health"); LocalAddress is the
// dial-able host:port used for grpcurl calls.
type GrpcServiceInfo struct {
	DisplayName  string `json:"displayName"`
	ServiceName  string `json:"serviceName"`
	LocalAddress string `json:"localAddress"`
	Namespace    string `json:"namespace"`
	K8sService   string `json:"k8sService"`
	ViaNodePort  bool   `json:"viaNodePort"`
}

// MethodInfo describes a single gRPC method, including its request/response
// proto type names. Returned by DescribeGrpcMethod.
type MethodInfo struct {
	Name         string `json:"name"`
	RequestType  string `json:"requestType"`
	ResponseType string `json:"responseType"`
}

// activeForward tracks a port-forward we own. ViaNodePort entries have pf=nil.
type activeForward struct {
	pf           *kubernetes.PortForward
	localAddress string
	viaNodePort  bool
}

// App holds runtime state shared across Wails-bound methods.
type App struct {
	ctx context.Context

	mu                sync.Mutex
	clusterDiscovery  *kubernetes.ClusterDiscovery
	currentClientset  *k8s.Clientset
	currentRestConfig *rest.Config
	currentContext    string
	accessibleNs      []string
	cfg               config.Config

	allocator *kubernetes.PortAllocator
	forwards  []*activeForward

	// hostMode is true when the user connected to the synthetic host-machine
	// environment instead of a real kubeconfig cluster. In host mode, all
	// k8s machinery (clientset, port-forwards, namespace probing) is bypassed.
	hostMode bool

	token       auth.TokenSet
	tokenUser   string
	refreshStop chan struct{}
}

// NewApp creates a new App. The wails runtime will call Startup once the
// frontend is ready.
func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[NewApp] Failed to load config, using defaults: %v", err)
		cfg = config.Default()
	}
	return &App{
		cfg:       cfg,
		allocator: kubernetes.NewPortAllocator(cfg.PortRangeStart, cfg.PortRangeEnd),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("[startup] Loaded config: namespaces=%v grpcPorts=%v portRange=%d-%d",
		a.cfg.Namespaces, a.cfg.GrpcPorts, a.cfg.PortRangeStart, a.cfg.PortRangeEnd)
}

// shutdown is invoked by Wails when the window is closing. It tears down any
// running port-forwards so kubectl child processes don't leak.
func (a *App) shutdown(ctx context.Context) {
	log.Println("[shutdown] Cleaning up port-forwards and refresh scheduler before exit")
	a.cancelRefreshLocked()
	a.stopAllForwards()
}

// GetClusters returns all clusters discovered in the user's kubeconfig, plus
// a synthetic "Host Machine" entry if any of the configured gRPC ports is
// currently listening on localhost.
func (a *App) GetClusters() ([]ClusterInfo, error) {
	log.Println("[GetClusters] Discovering clusters from kubeconfig")
	cd := kubernetes.NewClusterDiscovery("")
	clusters, err := cd.DiscoverClusters()
	if err != nil {
		log.Printf("[GetClusters] kubeconfig error: %v", err)
		// Don't fail the whole call — host-only environments are still useful.
		clusters = nil
	}

	a.mu.Lock()
	a.clusterDiscovery = cd
	a.mu.Unlock()

	out := make([]ClusterInfo, 0, len(clusters)+1)
	for _, c := range clusters {
		out = append(out, ClusterInfo{Name: c.Name, Context: c.Context, Server: c.Server})
		log.Printf("[GetClusters] %s (context=%s server=%s)", c.Name, c.Context, c.Server)
	}

	if isHostPortListening(hostProbePort) {
		out = append(out, ClusterInfo{
			Name:    hostDisplayName,
			Context: hostContextID,
			Server:  fmt.Sprintf("127.0.0.1:%d", hostProbePort),
		})
		log.Printf("[GetClusters] localhost detected on port %d", hostProbePort)
	} else {
		log.Printf("[GetClusters] No gRPC service listening on 127.0.0.1:%d", hostProbePort)
	}
	return out, nil
}

// isHostPortListening returns true if something is currently accepting TCP
// connections on 127.0.0.1:<port>. Short timeout — we don't want a slow
// probe to delay startup.
func isHostPortListening(port int) bool {
	if port <= 0 {
		return false
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ConnectToCluster builds a clientset for the chosen context and probes the
// configured namespaces. The returned slice contains one entry per configured
// namespace with its access status; callers can filter to Allowed=true.
//
// The synthetic hostContextID short-circuits this entire flow: there's no
// kubeconfig to read, so we just record the host-mode flag and return a
// single allowed pseudo-namespace.
func (a *App) ConnectToCluster(contextName string) ([]NamespaceInfo, error) {
	log.Printf("[ConnectToCluster] Connecting to context %q", contextName)

	if contextName == hostContextID {
		a.mu.Lock()
		a.hostMode = true
		a.currentClientset = nil
		a.currentRestConfig = nil
		a.currentContext = hostContextID
		a.accessibleNs = []string{hostNamespace}
		a.mu.Unlock()
		log.Println("[ConnectToCluster] Host machine mode active (kubeconfig bypassed)")
		return []NamespaceInfo{{Name: hostNamespace, Allowed: true}}, nil
	}

	cd := kubernetes.NewClusterDiscovery("")
	clientset, err := cd.GetClient(contextName)
	if err != nil {
		log.Printf("[ConnectToCluster] GetClient: %v", err)
		return nil, err
	}
	restCfg, err := cd.GetRESTConfig(contextName)
	if err != nil {
		log.Printf("[ConnectToCluster] GetRESTConfig: %v", err)
		return nil, err
	}

	a.mu.Lock()
	a.clusterDiscovery = cd
	a.currentClientset = clientset
	a.currentRestConfig = restCfg
	a.currentContext = contextName
	a.hostMode = false
	configured := append([]string(nil), a.cfg.Namespaces...)
	a.mu.Unlock()

	results := kubernetes.CheckNamespaceAccess(clientset, configured)

	infos := make([]NamespaceInfo, 0, len(results))
	allowed := make([]string, 0, len(results))
	for _, r := range results {
		infos = append(infos, NamespaceInfo{Name: r.Namespace, Allowed: r.Allowed, Reason: r.Reason})
		if r.Allowed {
			allowed = append(allowed, r.Namespace)
			log.Printf("[ConnectToCluster] Namespace allowed: %s", r.Namespace)
		} else {
			log.Printf("[ConnectToCluster] Namespace denied: %s (%s)", r.Namespace, r.Reason)
		}
	}

	a.mu.Lock()
	a.accessibleNs = allowed
	a.mu.Unlock()

	log.Printf("[ConnectToCluster] Connected. accessible=%d/%d namespaces", len(allowed), len(configured))
	return infos, nil
}

// DisconnectFromCluster tears down any active port-forwards and clears cached
// cluster state.
func (a *App) DisconnectFromCluster() error {
	a.stopAllForwards()
	a.mu.Lock()
	a.currentClientset = nil
	a.currentRestConfig = nil
	a.currentContext = ""
	a.accessibleNs = nil
	a.hostMode = false
	a.mu.Unlock()
	log.Println("[DisconnectFromCluster] Disconnected")
	return nil
}

// IsConnected returns whether a cluster client is currently active.
// Host-machine mode is also "connected" — it just doesn't have a clientset.
func (a *App) IsConnected() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentClientset != nil || a.hostMode
}

// GetAccessibleNamespaces returns the cached list of namespaces the user can
// access in the currently connected cluster.
func (a *App) GetAccessibleNamespaces() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.accessibleNs...)
}

// GetSettings returns the current persisted configuration.
func (a *App) GetSettings() config.Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg
}

// SaveSettings persists a new configuration to disk and updates the in-memory
// copy. The port allocator is rebuilt if the range changed.
func (a *App) SaveSettings(cfg config.Config) error {
	if cfg.PortRangeStart > 0 && cfg.PortRangeEnd > 0 && cfg.PortRangeEnd < cfg.PortRangeStart {
		return fmt.Errorf("portRangeEnd (%d) must be >= portRangeStart (%d)", cfg.PortRangeEnd, cfg.PortRangeStart)
	}
	if err := config.Save(cfg); err != nil {
		log.Printf("[SaveSettings] Save failed: %v", err)
		return err
	}
	a.mu.Lock()
	rangeChanged := cfg.PortRangeStart != a.cfg.PortRangeStart || cfg.PortRangeEnd != a.cfg.PortRangeEnd
	a.cfg = cfg
	if rangeChanged {
		a.allocator = kubernetes.NewPortAllocator(cfg.PortRangeStart, cfg.PortRangeEnd)
	}
	a.mu.Unlock()
	log.Printf("[SaveSettings] Saved. namespaces=%v grpcPorts=%v portRange=%d-%d",
		cfg.Namespaces, cfg.GrpcPorts, cfg.PortRangeStart, cfg.PortRangeEnd)
	return nil
}

// SelectNamespace tears down the previous namespace's port-forwards, then
// discovers gRPC services in the requested namespace. For each service it
// either uses the configured NodePort host:port directly, or starts a
// `kubectl port-forward` to a Running pod and reflects on the resulting local
// address. Reflection failures are logged but don't abort the entire scan.
func (a *App) SelectNamespace(namespace string) ([]GrpcServiceInfo, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	a.mu.Lock()
	clientset := a.currentClientset
	cfg := a.cfg
	allocator := a.allocator
	hostMode := a.hostMode
	a.mu.Unlock()

	if hostMode {
		return a.discoverHostMachineServices(cfg)
	}

	if clientset == nil {
		return nil, fmt.Errorf("not connected to a cluster")
	}

	// Always start fresh — old forwards from a different namespace are no
	// longer relevant.
	a.stopAllForwards()

	log.Printf("[SelectNamespace] Discovering gRPC services in %q (ports=%v)", namespace, cfg.GrpcPorts)
	disco := kubernetes.NewServiceDiscovery(clientset, cfg.GrpcPorts)
	services, err := disco.DiscoverGrpcServices(namespace)
	if err != nil {
		return nil, fmt.Errorf("discovering services: %w", err)
	}
	log.Printf("[SelectNamespace] Found %d candidate gRPC services", len(services))

	out := make([]GrpcServiceInfo, 0, len(services))
	for _, svc := range services {
		if pattern, matchedOn, skip := matchExcludeAny(cfg.ServiceExcludePatterns, svc.Name, svc.AppName); skip {
			log.Printf("[SelectNamespace] %s/%s: excluded by pattern %q (matched %s)", svc.Namespace, svc.Name, pattern, matchedOn)
			continue
		}
		var (
			localAddress string
			viaNodePort  bool
			pf           *kubernetes.PortForward
		)

		if svc.NodePort > 0 {
			localAddress = fmt.Sprintf("%s:%d", cfg.NodePortHost, svc.NodePort)
			viaNodePort = true
			log.Printf("[SelectNamespace] %s/%s: using NodePort %s", svc.Namespace, svc.Name, localAddress)
		} else {
			pods, err := disco.GetServicePods(svc.Namespace, svc.Selector)
			if err != nil {
				log.Printf("[SelectNamespace] %s/%s: list pods failed: %v", svc.Namespace, svc.Name, err)
				continue
			}
			if len(pods) == 0 {
				log.Printf("[SelectNamespace] %s/%s: no Running pods, skipping", svc.Namespace, svc.Name)
				continue
			}
			pod := pods[0]
			pf = kubernetes.NewPortForward(allocator, svc.Namespace, pod.Name, int(svc.Port))
			if err := pf.Start(); err != nil {
				log.Printf("[SelectNamespace] %s/%s: port-forward failed: %v", svc.Namespace, svc.Name, err)
				continue
			}
			localAddress = pf.GetLocalAddress()
		}

		reflected, err := grpcrefl.ListServices(localAddress)
		if err != nil {
			log.Printf("[SelectNamespace] %s/%s: reflection failed on %s: %v", svc.Namespace, svc.Name, localAddress, err)
			if pf != nil {
				pf.Stop()
			}
			continue
		}
		usable := filterUsableServices(reflected)
		if len(usable) == 0 {
			log.Printf("[SelectNamespace] %s/%s: no usable gRPC services on %s (after filtering health/reflection)", svc.Namespace, svc.Name, localAddress)
			if pf != nil {
				pf.Stop()
			}
			continue
		}

		af := &activeForward{pf: pf, localAddress: localAddress, viaNodePort: viaNodePort}
		a.mu.Lock()
		a.forwards = append(a.forwards, af)
		a.mu.Unlock()

		port := portFromAddress(localAddress)
		multiple := len(usable) > 1
		for _, serviceName := range usable {
			display := fmt.Sprintf("%s:%s", svc.AppName, port)
			if multiple {
				display = fmt.Sprintf("%s:%s · %s", svc.AppName, port, shortServiceName(serviceName))
			}
			out = append(out, GrpcServiceInfo{
				DisplayName:  display,
				ServiceName:  serviceName,
				LocalAddress: localAddress,
				Namespace:    svc.Namespace,
				K8sService:   svc.Name,
				ViaNodePort:  viaNodePort,
			})
			log.Printf("[SelectNamespace] %s -> %s (%s)", display, serviceName, localAddress)
		}
	}

	return out, nil
}

// discoverHostMachineServices probes the host gRPC port on 127.0.0.1 and runs
// server reflection. No port-forwarding is involved — the user's local
// process already provides the address.
func (a *App) discoverHostMachineServices(cfg config.Config) ([]GrpcServiceInfo, error) {
	a.stopAllForwards()

	if !isHostPortListening(hostProbePort) {
		log.Printf("[SelectNamespace] localhost: nothing listening on 127.0.0.1:%d", hostProbePort)
		return []GrpcServiceInfo{}, nil
	}

	localAddress := fmt.Sprintf("127.0.0.1:%d", hostProbePort)
	log.Printf("[SelectNamespace] localhost: reflecting %s", localAddress)
	reflected, err := grpcrefl.ListServices(localAddress)
	if err != nil {
		log.Printf("[SelectNamespace] localhost: reflection failed on %s: %v", localAddress, err)
		return []GrpcServiceInfo{}, nil
	}
	usable := filterUsableServices(reflected)
	if len(usable) == 0 {
		log.Printf("[SelectNamespace] localhost: no usable services on %s after filtering", localAddress)
		return []GrpcServiceInfo{}, nil
	}

	out := make([]GrpcServiceInfo, 0, len(usable))
	multiple := len(usable) > 1
	for _, serviceName := range usable {
		appName := shortServiceName(serviceName)
		display := fmt.Sprintf("%s:%d", appName, hostProbePort)
		if multiple {
			display = fmt.Sprintf("localhost:%d · %s", hostProbePort, appName)
		}
		out = append(out, GrpcServiceInfo{
			DisplayName:  display,
			ServiceName:  serviceName,
			LocalAddress: localAddress,
			Namespace:    hostNamespace,
			K8sService:   "",
			ViaNodePort:  false,
		})
		log.Printf("[SelectNamespace] %s -> %s (%s)", display, serviceName, localAddress)
	}
	return out, nil
}

func (a *App) stopAllForwards() {
	a.mu.Lock()
	forwards := a.forwards
	a.forwards = nil
	a.mu.Unlock()

	for _, af := range forwards {
		if af.pf != nil {
			af.pf.Stop()
		}
	}
}

// filterUsableServices removes reflection and standard health services that
// aren't meaningful to call from the UI. Reflection services are already
// stripped by ListServices, but we belt-and-suspenders here in case the upstream
// helper ever changes.
func filterUsableServices(services []string) []string {
	out := make([]string, 0, len(services))
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		switch s {
		case "grpc.health.v1.Health",
			"grpc.reflection.v1.ServerReflection",
			"grpc.reflection.v1alpha.ServerReflection":
			continue
		}
		out = append(out, s)
	}
	return out
}

// shortServiceName returns the last dot-separated segment of a fully-qualified
// gRPC service name (e.g. "solvasam.domaindata.v1.DomainDataService" → "DomainDataService").
func shortServiceName(serviceName string) string {
	if idx := strings.LastIndex(serviceName, "."); idx >= 0 {
		return serviceName[idx+1:]
	}
	return serviceName
}

// GetGrpcMethods lists method names available on the given gRPC service via
// server reflection.
func (a *App) GetGrpcMethods(localAddress, serviceName string) ([]string, error) {
	if localAddress == "" || serviceName == "" {
		return nil, fmt.Errorf("localAddress and serviceName are required")
	}
	methods, err := grpcrefl.GetGrpcMethods(localAddress, serviceName)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// DescribeGrpcMethod returns the request and response proto type names for a
// single method.
func (a *App) DescribeGrpcMethod(localAddress, serviceName, methodName string) (*MethodInfo, error) {
	if localAddress == "" || serviceName == "" || methodName == "" {
		return nil, fmt.Errorf("localAddress, serviceName, and methodName are required")
	}
	desc, err := grpcrefl.DescribeMethod(localAddress, serviceName, methodName)
	if err != nil {
		return nil, err
	}
	return &MethodInfo{
		Name:         desc.Name,
		RequestType:  desc.RequestType,
		ResponseType: desc.ResponseType,
	}, nil
}

// GenerateRequestSkeleton returns a JSON skeleton for the given proto message
// type, suitable for prefilling the request body editor.
func (a *App) GenerateRequestSkeleton(localAddress, requestType string) (string, error) {
	if localAddress == "" || requestType == "" {
		return "", fmt.Errorf("localAddress and requestType are required")
	}
	return grpcrefl.GenerateJsonSkeleton(localAddress, requestType)
}

// SendGrpcRequest invokes the given fully-qualified gRPC method with the
// supplied JSON request body via grpcurl, optionally attaching a bearer token.
// Returns the response JSON on success or a parsed gRPC error on failure.
func (a *App) SendGrpcRequest(localAddress, serviceName, methodName, requestBody, bearerToken string) (string, error) {
	if localAddress == "" || serviceName == "" || methodName == "" {
		return "", fmt.Errorf("localAddress, serviceName, and methodName are required")
	}
	if strings.TrimSpace(requestBody) == "" {
		requestBody = "{}"
	}
	return grpcrefl.SendGrpcRequest(localAddress, serviceName, methodName, requestBody, bearerToken)
}

// GenerateSampleRequest returns a JSON request body where zero-valued
// primitives in the skeleton are replaced with random sample values
// (random strings, random ints, random bools). Enum-looking strings ending
// in `_UNSPECIFIED` are left as-is.
func (a *App) GenerateSampleRequest(localAddress, requestType string) (string, error) {
	if localAddress == "" || requestType == "" {
		return "", fmt.Errorf("localAddress and requestType are required")
	}
	skel, err := grpcrefl.GenerateJsonSkeleton(localAddress, requestType)
	if err != nil {
		return "", err
	}
	return grpcrefl.RandomizeSkeleton(skel)
}

// AuthState is the per-call snapshot of token state surfaced to the frontend.
type AuthState struct {
	LoggedIn         bool   `json:"loggedIn"`
	Username         string `json:"username"`
	AccessToken      string `json:"accessToken"`
	ExpiresAt        int64  `json:"expiresAt"`        // unix seconds
	RefreshExpiresAt int64  `json:"refreshExpiresAt"` // unix seconds, 0 if unknown
}

func (a *App) buildAuthStateLocked() AuthState {
	state := AuthState{
		LoggedIn:    a.token.AccessToken != "",
		Username:    a.tokenUser,
		AccessToken: a.token.AccessToken,
	}
	if !a.token.ObtainedAt.IsZero() {
		state.ExpiresAt = a.token.AccessExpiresAt().Unix()
		if rx := a.token.RefreshExpiresAt(); !rx.IsZero() {
			state.RefreshExpiresAt = rx.Unix()
		}
	}
	return state
}

// GetAuthState returns the current login snapshot. The frontend calls this on
// startup to rehydrate after a hot reload.
func (a *App) GetAuthState() AuthState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.buildAuthStateLocked()
}

// Login performs an OIDC password-grant against the configured token endpoint
// and starts the refresh scheduler. The returned AuthState reflects the new
// token.
func (a *App) Login(username, password string) (AuthState, error) {
	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()

	log.Printf("[Login] Requesting token for user=%q endpoint=%s", username, cfg.TokenEndpoint)
	ts, err := auth.Login(cfg.TokenEndpoint, cfg.ClientID, username, password)
	if err != nil {
		log.Printf("[Login] Failed: %v", err)
		return AuthState{}, err
	}
	a.applyToken(ts, username, "login")
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.buildAuthStateLocked(), nil
}

// RefreshToken trades the current refresh token for a new access token on
// demand. Used by the manual refresh button in the UI.
func (a *App) RefreshToken() (AuthState, error) {
	a.mu.Lock()
	cfg := a.cfg
	rt := a.token.RefreshToken
	user := a.tokenUser
	a.mu.Unlock()

	if rt == "" {
		return AuthState{}, fmt.Errorf("no refresh token available — please log in again")
	}
	log.Printf("[RefreshToken] Refreshing token for user=%q", user)
	ts, err := auth.Refresh(cfg.TokenEndpoint, cfg.ClientID, rt)
	if err != nil {
		log.Printf("[RefreshToken] Failed: %v", err)
		return AuthState{}, err
	}
	a.applyToken(ts, user, "manual-refresh")
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.buildAuthStateLocked(), nil
}

// Logout clears the in-memory token and cancels the refresh scheduler. The UI
// is responsible for closing any token-dependent panels.
func (a *App) Logout() {
	log.Println("[Logout] Clearing token")
	a.cancelRefreshLocked()
	a.mu.Lock()
	a.token = auth.TokenSet{}
	a.tokenUser = ""
	state := a.buildAuthStateLocked()
	a.mu.Unlock()
	a.emitAuthEvent("token:cleared", state)
}

// applyToken stores a new TokenSet, restarts the refresh scheduler, and emits
// a "token:refreshed" event so the frontend can update its store.
func (a *App) applyToken(ts auth.TokenSet, username, source string) {
	a.cancelRefreshLocked()

	a.mu.Lock()
	a.token = ts
	if username != "" {
		a.tokenUser = username
	}
	state := a.buildAuthStateLocked()
	a.mu.Unlock()

	log.Printf("[applyToken] (%s) accessExpiresAt=%s refreshExpiresAt=%s",
		source, ts.AccessExpiresAt().Format(time.RFC3339), ts.RefreshExpiresAt().Format(time.RFC3339))

	a.emitAuthEvent("token:refreshed", state)
	a.scheduleRefresh(ts.ExpiresIn)
}

// scheduleRefresh starts a goroutine that fires `Refresh` 30s before access
// token expiry. If refresh fails it emits "token:expired" and gives up.
func (a *App) scheduleRefresh(expiresIn int) {
	if expiresIn <= 0 {
		log.Println("[scheduleRefresh] expiresIn <= 0; skipping auto-refresh")
		return
	}
	delay := time.Duration(expiresIn-30) * time.Second
	if delay < 5*time.Second {
		delay = 5 * time.Second
	}

	stop := make(chan struct{})
	a.mu.Lock()
	a.refreshStop = stop
	a.mu.Unlock()

	go func() {
		log.Printf("[scheduleRefresh] sleeping %s before next refresh", delay)
		select {
		case <-time.After(delay):
		case <-stop:
			log.Println("[scheduleRefresh] cancelled before fire")
			return
		}

		a.mu.Lock()
		// Bail out if a manual refresh/login replaced this scheduler in the meantime.
		if a.refreshStop != stop {
			a.mu.Unlock()
			return
		}
		cfg := a.cfg
		rt := a.token.RefreshToken
		user := a.tokenUser
		refreshExp := a.token.RefreshExpiresAt()
		a.mu.Unlock()

		if rt == "" {
			log.Println("[scheduleRefresh] no refresh token; cannot continue")
			a.emitAuthEvent("token:expired", AuthState{Username: user})
			return
		}
		if !refreshExp.IsZero() && time.Now().After(refreshExp) {
			log.Println("[scheduleRefresh] refresh token already expired")
			a.emitAuthEvent("token:expired", AuthState{Username: user})
			return
		}

		ts, err := auth.Refresh(cfg.TokenEndpoint, cfg.ClientID, rt)
		if err != nil {
			log.Printf("[scheduleRefresh] refresh failed: %v", err)
			a.emitAuthEvent("token:expired", AuthState{Username: user})
			return
		}
		a.applyToken(ts, user, "auto-refresh")
	}()
}

// cancelRefreshLocked stops any in-flight refresh goroutine. Safe to call from
// any state.
func (a *App) cancelRefreshLocked() {
	a.mu.Lock()
	stop := a.refreshStop
	a.refreshStop = nil
	a.mu.Unlock()
	if stop != nil {
		close(stop)
	}
}

func (a *App) emitAuthEvent(event string, state AuthState) {
	if a.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(a.ctx, event, state)
}

// matchExcludeAny tries every pattern against every candidate name. It returns
// the matching pattern, a label identifying which candidate matched (for log
// readability), and true if any candidate hit. Empty candidates and patterns
// are skipped. Invalid glob patterns fall back to substring match so a typo'd
// entry still does *something* useful instead of silently failing.
func matchExcludeAny(patterns []string, candidates ...string) (string, string, bool) {
	labels := []string{"name", "appName", "extra"}
	for i, candidate := range candidates {
		if candidate == "" {
			continue
		}
		lname := strings.ToLower(candidate)
		for _, p := range patterns {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			lp := strings.ToLower(p)
			matched, err := path.Match(lp, lname)
			if err != nil {
				if strings.Contains(lname, strings.Trim(lp, "*")) {
					return p, label(labels, i), true
				}
				continue
			}
			if matched {
				return p, label(labels, i), true
			}
		}
	}
	return "", "", false
}

func label(labels []string, i int) string {
	if i < len(labels) {
		return labels[i]
	}
	return fmt.Sprintf("candidate[%d]", i)
}

func portFromAddress(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[i+1:]
		}
	}
	return addr
}
