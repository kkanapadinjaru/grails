package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
)

// AuthEndpoint binds a (cluster, namespace) pair to a Keycloak (or compatible
// OIDC) configuration. "*" in either field is a wildcard matching anything.
// Lookup order is exact → cluster-wildcard → namespace-wildcard → (*,*).
//
// TokenURL may contain "{realm}", which is substituted from the resolver
// response when present. RealmResolverURL may contain "{subdomain}", which
// is substituted from the user's input in the login modal. RealmJSONPath is
// a dot path into the resolver's JSON response (e.g. "realm" or
// "data.realm"). Leave RealmResolverURL empty when TokenURL is fully
// qualified (no {realm} placeholder).
type AuthEndpoint struct {
	Cluster          string `json:"cluster"`
	Namespace        string `json:"namespace"`
	TokenURL         string `json:"tokenUrl"`
	RealmResolverURL string `json:"realmResolverUrl"`
	RealmJSONPath    string `json:"realmJsonPath"`
}

// Config is the user-editable application configuration persisted to
// %APPDATA%/grails/config.json (or the OS equivalent).
type Config struct {
	Namespaces             []string       `json:"namespaces"`
	PortRangeStart         int            `json:"portRangeStart"`
	PortRangeEnd           int            `json:"portRangeEnd"`
	GrpcPorts              []int          `json:"grpcPorts"`
	DiscoveryConcurrency   int            `json:"discoveryConcurrency"`
	NodePortHost           string         `json:"nodePortHost"`
	AuthProvider           string         `json:"authProvider"`
	ClientID               string         `json:"clientId"`
	AuthEndpoints          []AuthEndpoint `json:"authEndpoints"`
	ServiceExcludePatterns []string       `json:"serviceExcludePatterns"`

	// ParentClaimMap maps a URL-pattern segment prefix (the literal token
	// before "/*" in a (google.api.http) parent binding) to a JWT claim
	// name. When a method's request type carries a `parent` binding like
	// `{parent=o/*}`, the sample-request scaffolder substitutes the value
	// of the mapped claim ("owner_id" by default) for the wildcard. Empty
	// or unmapped prefixes fall back to the random sample value.
	ParentClaimMap map[string]string `json:"parentClaimMap"`
}

// Default returns a Config initialized with the documented defaults.
func Default() Config {
	return Config{
		Namespaces:             []string{"default", "am-dev", "am-qa", "am-demo"},
		PortRangeStart:         35000,
		PortRangeEnd:           60000,
		GrpcPorts:              []int{5001, 5002},
		DiscoveryConcurrency:   5,
		NodePortHost:           "127.0.0.1",
		AuthProvider:           "keycloak",
		ClientID:               "",
		AuthEndpoints:          []AuthEndpoint{},
		ServiceExcludePatterns: []string{"*wassups", "*-lb"},
		ParentClaimMap:         map[string]string{"o": "owner_id"},
	}
}

// MatchAuthEndpoint returns the most specific endpoint configured for the
// given (cluster, namespace) pair. Lookup order: exact, cluster-wildcard,
// namespace-wildcard, total-wildcard. Returns false if nothing matches.
func (c Config) MatchAuthEndpoint(cluster, namespace string) (AuthEndpoint, bool) {
	tiers := [][2]string{
		{cluster, namespace},
		{cluster, "*"},
		{"*", namespace},
		{"*", "*"},
	}
	for _, t := range tiers {
		for _, ep := range c.AuthEndpoints {
			if ep.Cluster == t[0] && ep.Namespace == t[1] {
				return ep, true
			}
		}
	}
	return AuthEndpoint{}, false
}

// configPath returns the absolute path to the config file.
func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config dir: %w", err)
	}
	return filepath.Join(dir, "grails", "config.json"), nil
}

// Load reads the config from disk, creating it with defaults if missing.
func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if saveErr := Save(cfg); saveErr != nil {
				log.Printf("[config.Load] Could not seed default config at %s: %v", path, saveErr)
			} else {
				log.Printf("[config.Load] Created default config at %s", path)
			}
			return cfg, nil
		}
		return Default(), fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("parsing config %s: %w", path, err)
	}

	// Backfill any missing fields from defaults so old config files keep working.
	d := Default()
	if len(cfg.Namespaces) == 0 {
		cfg.Namespaces = d.Namespaces
	}
	if cfg.PortRangeStart == 0 {
		cfg.PortRangeStart = d.PortRangeStart
	}
	if cfg.PortRangeEnd == 0 {
		cfg.PortRangeEnd = d.PortRangeEnd
	}
	if len(cfg.GrpcPorts) == 0 {
		cfg.GrpcPorts = d.GrpcPorts
	}
	if cfg.DiscoveryConcurrency <= 0 {
		cfg.DiscoveryConcurrency = d.DiscoveryConcurrency
	}
	if cfg.NodePortHost == "" {
		cfg.NodePortHost = d.NodePortHost
	}
	if cfg.AuthProvider == "" {
		cfg.AuthProvider = d.AuthProvider
	}
	if cfg.AuthEndpoints == nil {
		cfg.AuthEndpoints = []AuthEndpoint{}
	}
	if cfg.ParentClaimMap == nil {
		cfg.ParentClaimMap = d.ParentClaimMap
	}

	// Merge default exclude patterns so protective heuristics added in code
	// reach existing users without requiring manual Profile edits. In-memory
	// only — we don't rewrite the file behind the user's back.
	for _, p := range d.ServiceExcludePatterns {
		if !slices.Contains(cfg.ServiceExcludePatterns, p) {
			cfg.ServiceExcludePatterns = append(cfg.ServiceExcludePatterns, p)
		}
	}

	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	log.Printf("[config.Save] Wrote config to %s", path)
	return nil
}
