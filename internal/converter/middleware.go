package converter

import (
	"fmt"

	"github.com/databotic/traefik-migration-tool/internal/utils"
	containous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type MiddleWare struct{}

func NewMiddleWare() *MiddleWare {
	return &MiddleWare{}
}

func (m *MiddleWare) Transform(object runtime.Object) (runtime.Object, error) {
	v2MiddleWare, ok := object.(*containous.Middleware)
	if !ok {
		return nil, fmt.Errorf("err")
	}

	if utils.HasDepricatedMiddleWareOptions(v2MiddleWare) {
		fmt.Printf("middleware %s has depricated options and it should be fixed manually\n", v2MiddleWare.Name)
	}

	middleware := &traefikio.Middleware{
		TypeMeta: v1.TypeMeta{Kind: v2MiddleWare.Kind, APIVersion: utils.APIVersion},
		ObjectMeta: v1.ObjectMeta{
			Name: v2MiddleWare.ObjectMeta.Name, Namespace: v2MiddleWare.ObjectMeta.Namespace,
			Annotations: utils.FilterAnnotations(v2MiddleWare.Annotations), Labels: v2MiddleWare.Labels,
		},
	}

	spec, err := utils.AsType[traefikio.MiddlewareSpec](v2MiddleWare.Spec)
	if err != nil {
		return nil, fmt.Errorf("error converting middleware to v3 %v", err)
	}

	if spec.IPWhiteList != nil {
		ipAllowList, err := utils.AsType[dynamic.IPAllowList](v2MiddleWare.Spec.IPWhiteList)
		if err != nil {
			fmt.Println("error converting IPWhiteList to IPAllowList")
		}
		spec.IPWhiteList = nil
		spec.IPAllowList = ipAllowList
	}

	middleware.Spec = *spec
	return middleware, nil
}
