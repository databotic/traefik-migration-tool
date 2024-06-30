package resources

import (
	"context"
	"fmt"

	"github.com/databotic/traefik-migration-tool/internal/utils"
	containous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	containous_v1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	traefikio_v1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MiddleWare struct {
	DryRun       []string
	Namespace    string
	ResourceName string

	ContainousClient *containous.Clientset
	TraefikioClient  *traefikio.Clientset
}

func InitiliazeMiddleWare(config *ResourceInput) (*MiddleWare, error) {
	var dryRun []string
	if config.DryRun {
		dryRun = []string{"ALL"}
	}

	return &MiddleWare{
		DryRun:       dryRun,
		Namespace:    config.Namespace,
		ResourceName: config.ResourceName,

		ContainousClient: config.ContainousClient,
		TraefikioClient:  config.TraefikioClient,
	}, nil
}

func (m *MiddleWare) GetMiddleWares() ([]containous_v1alpha1.Middleware, error) {
	request := m.ContainousClient.TraefikContainousV1alpha1().Middlewares(m.Namespace)

	if m.ResourceName != "" {
		middleware, err := request.Get(context.TODO(), m.ResourceName, v1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return []containous_v1alpha1.Middleware{*middleware}, nil
	}

	middleware, err := request.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return middleware.Items, err
}

func (m *MiddleWare) Migrate() error {
	v2MiddleWares, err := m.GetMiddleWares()
	if err != nil {
		return err
	}

	for _, middleware := range v2MiddleWares {
		if utils.HasDepricatedMiddleWareOptions(&middleware) {
			fmt.Printf("middleware %s has depricated options and it should be migrated manually, skipping\n", middleware.Name)
			continue
		}
		v3Middleware, err := m.convert(middleware)
		if err != nil {
			fmt.Printf("error migrating middleware %s/%s: %v", middleware.Name, middleware.Namespace, err)
			continue
		}

		b, _ := utils.EncodeYaml(v3Middleware)
		fmt.Printf("%s", b)

		_, err = m.TraefikioClient.TraefikV1alpha1().Middlewares(v3Middleware.Namespace).Create(
			context.TODO(), v3Middleware, v1.CreateOptions{DryRun: m.DryRun},
		)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				fmt.Printf("middleware with name %s already exists in %s namespace",
					v3Middleware.Name, v3Middleware.Namespace,
				)
				continue
			}
			return err
		}
	}

	return nil
}

func (m *MiddleWare) convertIPWhiteList(o containous_v1alpha1.Middleware) (*traefikio_v1alpha1.Middleware, error) {
	ipAllowList, err := utils.AsType[dynamic.IPAllowList](o.Spec.IPWhiteList)
	if err != nil {
		return nil, fmt.Errorf("error converting IPWhitelist middleware to IPAllowList")
	}

	return &traefikio_v1alpha1.Middleware{
		TypeMeta: v1.TypeMeta{Kind: o.Kind, APIVersion: utils.APIVersion},
		ObjectMeta: v1.ObjectMeta{
			Name: o.ObjectMeta.Name, Namespace: o.ObjectMeta.Namespace,
			Annotations: utils.FilterAnnotations(o.Annotations), Labels: o.Labels,
		},
		Spec: traefikio_v1alpha1.MiddlewareSpec{
			IPAllowList: ipAllowList,
		},
	}, nil
}

func (m *MiddleWare) convert(o containous_v1alpha1.Middleware) (*traefikio_v1alpha1.Middleware, error) {
	if o.Spec.IPWhiteList != nil {
		return m.convertIPWhiteList(o)
	}

	middleware := &traefikio_v1alpha1.Middleware{
		TypeMeta: v1.TypeMeta{Kind: o.TypeMeta.Kind, APIVersion: utils.APIVersion},
		ObjectMeta: v1.ObjectMeta{
			Name: o.ObjectMeta.Name, Namespace: o.ObjectMeta.Namespace,
			Annotations: utils.FilterAnnotations(o.Annotations), Labels: o.Labels,
		},
	}
	spec, err := utils.AsType[traefikio_v1alpha1.MiddlewareSpec](o.Spec)
	if err != nil {
		return nil, err
	}
	middleware.Spec = *spec

	return middleware, nil
}
