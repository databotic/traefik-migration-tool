package parser

import (
	"fmt"
	"io"
	"os"

	containous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	containous.AddToScheme(scheme.Scheme)
	traefikio.AddToScheme(scheme.Scheme)

}

func ParseManifest(file *os.File) ([]runtime.Object, error) {
	defer func() { _ = file.Close() }()

	decoder := yaml.NewYAMLOrJSONDecoder(file, 4096)
	deser := scheme.Codecs.UniversalDeserializer()

	var objects []runtime.Object

	for {
		ext := runtime.RawExtension{}
		if err := decoder.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error parsing manifest: %v", err)
		}

		obj, _, err := deser.Decode([]byte(ext.Raw), nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error parsing object in manifest: %v", err)
		}

		objects = append(objects, obj)
	}

	return objects, nil
}
