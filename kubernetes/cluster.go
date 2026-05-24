package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Cluster represents a Kubernetes cluster configuration
type Cluster struct {
	Name     string
	Context  string
	Server   string
	Location string
}

// ClusterDiscovery handles discovering clusters from kubeconfig
type ClusterDiscovery struct {
	kubeconfig string
}

// NewClusterDiscovery creates a new ClusterDiscovery instance. If kubeconfig is
// empty, the default location is used (~/.kube/config or %USERPROFILE%\.kube\config).
func NewClusterDiscovery(kubeconfig string) *ClusterDiscovery {
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	return &ClusterDiscovery{kubeconfig: kubeconfig}
}

// DiscoverClusters returns all available clusters from kubeconfig
func (cd *ClusterDiscovery) DiscoverClusters() ([]Cluster, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = cd.kubeconfig

	kubeConfig, err := loadingRules.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %w", err)
	}

	var clusters []Cluster
	for contextName, context := range kubeConfig.Contexts {
		if context.Cluster == "" {
			continue
		}

		clusterConfig, exists := kubeConfig.Clusters[context.Cluster]
		if !exists {
			continue
		}

		location := "other"
		if kubeConfig.CurrentContext == contextName {
			location = "current"
		}

		clusters = append(clusters, Cluster{
			Name:     context.Cluster,
			Context:  contextName,
			Server:   clusterConfig.Server,
			Location: location,
		})
	}

	return clusters, nil
}

// GetClient creates a Kubernetes client for a specific context
func (cd *ClusterDiscovery) GetClient(contextName string) (*kubernetes.Clientset, error) {
	config, err := cd.GetRESTConfig(contextName)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %w", err)
	}
	return clientset, nil
}

// GetRESTConfig returns the REST config for a specific context
func (cd *ClusterDiscovery) GetRESTConfig(contextName string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = cd.kubeconfig

	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating REST config: %w", err)
	}
	return config, nil
}
