package resources

import (
	"context"
	"fmt"

	"github.com/databotic/traefik-migration-tool/internal/utils"
	containous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	containous_v1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	traefikio_v1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressRoute struct {
	DryRun       []string
	Namespace    string
	ResourceName string

	ContainousClient *containous.Clientset
	TraefikioClient  *traefikio.Clientset
}

func InitiliazeIngressRoute(config *ResourceInput) (*IngressRoute, error) {
	var dryRun []string
	if config.DryRun {
		dryRun = []string{"ALL"}
	}

	return &IngressRoute{
		DryRun:       dryRun,
		Namespace:    config.Namespace,
		ResourceName: config.ResourceName,

		ContainousClient: config.ContainousClient,
		TraefikioClient:  config.TraefikioClient,
	}, nil
}

func (m *IngressRoute) GetIngressRoutes() ([]containous_v1alpha1.IngressRoute, error) {
	request := m.ContainousClient.TraefikContainousV1alpha1().IngressRoutes(m.Namespace)

	if m.ResourceName != "" {
		ingressRoute, err := request.Get(context.TODO(), m.ResourceName, v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		return []containous_v1alpha1.IngressRoute{*ingressRoute}, nil
	}

	ingressRoutes, err := request.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return ingressRoutes.Items, err
}

func (m *IngressRoute) Migrate() error {
	v2IngressRoutes, err := m.GetIngressRoutes()
	if err != nil {
		return err
	}

	for _, v2IngressRoute := range v2IngressRoutes {
		v3IngressRoute, err := m.convert(v2IngressRoute)
		if err != nil {
			return err
		}
		_, err = m.TraefikioClient.TraefikV1alpha1().IngressRoutes(v3IngressRoute.Namespace).Create(
			context.TODO(), v3IngressRoute, v1.CreateOptions{DryRun: m.DryRun},
		)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				fmt.Printf("ingressroute with name %s already exists in %s namespace\n",
					v3IngressRoute.Name, v3IngressRoute.Namespace,
				)
				continue
			}
			return err
		}
	}
	return nil
}

func (m *IngressRoute) convert(v2 containous_v1alpha1.IngressRoute) (*traefikio_v1alpha1.IngressRoute, error) {
	ingressRoute := &traefikio_v1alpha1.IngressRoute{
		TypeMeta: v1.TypeMeta{Kind: "IngressRoute", APIVersion: utils.APIVersion},
		ObjectMeta: v1.ObjectMeta{
			Name: v2.ObjectMeta.Name, Namespace: v2.ObjectMeta.Namespace,
			Annotations: utils.FilterAnnotations(v2.Annotations), Labels: v2.Labels,
		},
	}

	tls, err := utils.AsType[traefikio_v1alpha1.TLS](v2.Spec.TLS)
	if err != nil {
		return nil, err
	}
	ingressRoute.Spec = traefikio_v1alpha1.IngressRouteSpec{
		EntryPoints: v2.Spec.EntryPoints, TLS: tls,
	}
	for _, v2Route := range v2.Spec.Routes {
		route := traefikio_v1alpha1.Route{
			Match: v2Route.Match, Kind: v2Route.Kind,
			Priority: v2Route.Priority, Syntax: "v2",
		}

		for _, v2Svc := range v2Route.Services {
			svc, err := utils.AsType[traefikio_v1alpha1.Service](v2Svc)
			if err != nil {
				return nil, err
			}
			route.Services = append(route.Services, *svc)
		}

		for _, v2MiddleWares := range v2Route.Middlewares {
			middlewares, err := utils.AsType[traefikio_v1alpha1.MiddlewareRef](v2MiddleWares)
			if err != nil {
				return nil, err
			}
			route.Middlewares = append(route.Middlewares, *middlewares)
		}
		ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, route)
	}
	return ingressRoute, nil
}
