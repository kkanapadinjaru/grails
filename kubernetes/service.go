package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GrpcService describes a Kubernetes Service whose spec exposes one of the
// configured gRPC ports.
type GrpcService struct {
	Name      string
	Namespace string
	AppName   string
	Port      int32
	NodePort  int32
	Selector  map[string]string
}

// ServiceDiscovery scans a single namespace for gRPC services.
type ServiceDiscovery struct {
	clientset *kubernetes.Clientset
	grpcPorts []int32
}

// NewServiceDiscovery builds a discoverer that matches Service ports against
// any of grpcPorts (defaulting to 5001/5002 if empty).
func NewServiceDiscovery(clientset *kubernetes.Clientset, grpcPorts []int) *ServiceDiscovery {
	ports := make([]int32, 0, len(grpcPorts))
	for _, p := range grpcPorts {
		ports = append(ports, int32(p))
	}
	if len(ports) == 0 {
		ports = []int32{5001, 5002}
	}
	return &ServiceDiscovery{clientset: clientset, grpcPorts: ports}
}

// DiscoverGrpcServices returns every Service in the given namespace whose
// ServicePort.Port matches one of the configured gRPC ports.
func (sd *ServiceDiscovery) DiscoverGrpcServices(namespace string) ([]GrpcService, error) {
	services, err := sd.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing services in %s: %w", namespace, err)
	}

	var out []GrpcService
	for _, svc := range services.Items {
		for _, port := range svc.Spec.Ports {
			if !sd.portMatches(port.Port) {
				continue
			}
			appName := svc.Labels["app.kubernetes.io/name"]
			if appName == "" {
				appName = svc.Name
			}
			out = append(out, GrpcService{
				Name:      svc.Name,
				Namespace: svc.Namespace,
				AppName:   appName,
				Port:      port.Port,
				NodePort:  port.NodePort,
				Selector:  svc.Spec.Selector,
			})
			break
		}
	}
	return out, nil
}

func (sd *ServiceDiscovery) portMatches(port int32) bool {
	for _, p := range sd.grpcPorts {
		if p == port {
			return true
		}
	}
	return false
}

// GetServicePods returns Running pods that match a service's selector.
func (sd *ServiceDiscovery) GetServicePods(namespace string, selector map[string]string) ([]corev1.Pod, error) {
	if len(selector) == 0 {
		return nil, nil
	}
	selectorString := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: selector})
	pods, err := sd.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectorString,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pods in %s: %w", namespace, err)
	}

	running := make([]corev1.Pod, 0, len(pods.Items))
	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodRunning {
			running = append(running, p)
		}
	}
	return running, nil
}
