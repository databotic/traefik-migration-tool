package converter

import (
	"regexp"

	"github.com/databotic/traefik-migration-tool/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	rulePattern  = regexp.MustCompile(`(\w+)\((` + "`[^`]*?(?:`,\\s*`[^`]+)*?`" + `)\)`)
	regexPattern = regexp.MustCompile(`[*+?()|{}[\]\\]`)
)

type ConvertFactory interface {
	Transform(object runtime.Object) (runtime.Object, error)
}

type Converter struct {
	converters map[string]ConvertFactory
}

func New() (*Converter, error) {
	IngressRoute, err := NewIngressRoute()
	if err != nil {
		return nil, err
	}

	converters := map[string]ConvertFactory{
		"IngressRoute.traefik.containo.us": IngressRoute,
		"Middleware.traefik.containo.us":   NewMiddleWare(),
	}

	return &Converter{converters: converters}, nil
}

func (c *Converter) Do(objects []runtime.Object) ([]runtime.Object, error) {
	var converted []runtime.Object

	for _, o := range objects {
		gvk := o.GetObjectKind()
		
		gk := gvk.GroupVersionKind().GroupKind().String()
		if converter, ok := c.converters[gk]; ok {
			object, err := converter.Transform(o)
			if err != nil {
				return converted, err
			}

			converted = append(converted, object)
			continue
		}
		converted = append(converted, o)
	}

	return converted, nil
}

func (c *Converter) EncodeYaml(object runtime.Object) ([]byte, error) {
	encoder := scheme.Codecs.EncoderForVersion(
		utils.YAMLCodec{}, object.GetObjectKind().GroupVersionKind().GroupVersion(),
	)
	return runtime.Encode(encoder, object)
}
