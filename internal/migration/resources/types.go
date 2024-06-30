package resources

import (
	containous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
)

type ResourceInput struct {
	DryRun       bool
	Namespace    string
	ResourceType string
	ResourceName string

	ContainousClient *containous.Clientset
	TraefikioClient  *traefikio.Clientset
}
