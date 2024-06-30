package migration

import (
	"fmt"
	"strings"

	"github.com/databotic/traefik-migration-tool/internal/migration/resources"
	containous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	"k8s.io/client-go/rest"
	ctrlconf "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Resource interface {
	Migrate() error
}

type Migration struct {
	KubeConfig       *rest.Config
	ContainousClient *containous.Clientset
	TraefikioClient  *traefikio.Clientset
}

func New() (*Migration, error) {
	kubeConfig, err := ctrlconf.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config: %w", err)
	}

	traefikioClientSet, err := traefikio.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error building traefik v3 clientset: %v", err)
	}

	containousClientSet, err := containous.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error building traefik v2 clientset: %v", err)
	}

	return &Migration{
		KubeConfig:       kubeConfig,
		ContainousClient: containousClientSet,
		TraefikioClient:  traefikioClientSet,
	}, nil
}

func (m *Migration) Run(resourceType, resourceName, namespace string, drynRun bool) error {
	resource, err := m.GetResource(resourceType, resourceName, namespace, drynRun)
	if err != nil {
		return err
	}

	return resource.Migrate()
}

func (m *Migration) GetResource(resourceType, resourceName, namespace string, drynRun bool) (Resource, error) {
	input := &resources.ResourceInput{
		DryRun:       drynRun,
		ResourceType: resourceType,
		ResourceName: resourceName,
		Namespace:    namespace,

		ContainousClient: m.ContainousClient,
		TraefikioClient:  m.TraefikioClient,
	}

	switch strings.ToLower(resourceType) {
	case "ingressroute":
		return resources.InitiliazeIngressRoute(input)
	case "middleware":
		return resources.InitiliazeMiddleWare(input)
	default:
		return nil, fmt.Errorf("resource type %s is not implemented", resourceType)
	}
}
