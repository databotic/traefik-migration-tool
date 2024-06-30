package utils

import (
	"io"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/test-infra/pkg/genyaml"
)

// custom codec to not output a creationTimestamp when marshalling as YAML and set the indentation properly
type YAMLCodec struct{}

func (YAMLCodec) Decode(_ []byte, gvk *schema.GroupVersionKind, obj runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	return obj, gvk, nil
}

func (YAMLCodec) Encode(obj runtime.Object, w io.Writer) error {
	yamlEncoder := yaml.NewEncoder(w)
	yamlEncoder.SetIndent(2)

	// this will take care of removing null filds whan marshalling.
	cm, err := genyaml.NewCommentMap(nil)
	if err != nil {
		return err
	}

	if err := cm.EncodeYaml(obj, yamlEncoder); err != nil {
		return errors.Wrap(err, "failed to marshal yaml manifest")
	}

	return nil
}

func (YAMLCodec) Identifier() runtime.Identifier {
	return runtime.Identifier("custom_codec")
}
