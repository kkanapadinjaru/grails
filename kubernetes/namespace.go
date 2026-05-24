package kubernetes

import (
	"context"
	"log"
	"time"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NamespaceCheckResult is the outcome of probing one namespace.
type NamespaceCheckResult struct {
	Namespace string
	Allowed   bool
	Reason    string
}

// CheckNamespaceAccess probes each configured namespace via SelfSubjectAccessReview
// for "list services" permission. Namespaces the user can't access are reported
// as denied, with a reason — the caller decides what to log.
func CheckNamespaceAccess(clientset *kubernetes.Clientset, configured []string) []NamespaceCheckResult {
	results := make([]NamespaceCheckResult, 0, len(configured))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, ns := range configured {
		review := &authv1.SelfSubjectAccessReview{
			Spec: authv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: ns,
					Verb:      "list",
					Group:     "",
					Resource:  "services",
				},
			},
		}

		resp, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
		if err != nil {
			log.Printf("[CheckNamespaceAccess] %s: SelfSubjectAccessReview failed: %v", ns, err)
			results = append(results, NamespaceCheckResult{Namespace: ns, Allowed: false, Reason: err.Error()})
			continue
		}

		if resp.Status.Allowed {
			results = append(results, NamespaceCheckResult{Namespace: ns, Allowed: true})
		} else {
			reason := resp.Status.Reason
			if reason == "" {
				reason = "forbidden"
			}
			log.Printf("[CheckNamespaceAccess] %s: denied (%s)", ns, reason)
			results = append(results, NamespaceCheckResult{Namespace: ns, Allowed: false, Reason: reason})
		}
	}

	return results
}
