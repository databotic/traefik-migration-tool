package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

var Version = "dev"

var RegexpCompileFunc = regexp.Compile

type regexpType int

const (
	RegexpTypePath regexpType = iota
	RegexpTypePrefix
	RegexpTypeQuery
	RegexpTypeHost
)

var AnnotationsTobeSkipped = []string{
	"kubectl.kubernetes.io/last-applied-configuration",
}

var APIVersion string = "traefik.io/v1alpha1"

var DepricatedMiddlewareOpts = map[string][]string{
	"Headers": []string{
		"SSLRedirect", "SSLTemporaryRedirect", "SSLHost",
		"SSLForceHost", "FeaturePolicy",
	},
	"StripPrefix": []string{"ForceSlash"},
}

func EncodeYaml(o interface{}) ([]byte, error) {
	var out bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&out)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(o); err != nil {
		return nil, errors.Wrap(err, "failed to marshal yaml manifest")
	}

	return out.Bytes(), nil
}

func ToUnstructured(o interface{}) (map[string]interface{}, error) {
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o)
	if err != nil {
		return nil, err
	}
	return unstructuredMap, nil
}

// copied from mux.Route
func RouteRegexp(tpl string, typ regexpType) (string, error) {
	idxs, errBraces := braceIndices(tpl)
	if errBraces != nil {
		return "", errBraces
	}
	// no need to modify if the string doesn't contain braces
	if len(idxs) == 0 {
		return tpl, nil
	}
	// Backup the original.
	template := tpl
	// Now let's parse it.
	defaultPattern := "[^/]+"
	if typ == RegexpTypeQuery {
		defaultPattern = ".*"
	} else if typ == RegexpTypeHost {
		defaultPattern = "[^.]+"
	}
	pattern := bytes.NewBufferString("")
	pattern.WriteByte('^')
	var end int
	for i := 0; i < len(idxs); i += 2 {
		// Set all values we are interested in.
		raw := tpl[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(tpl[idxs[i]+1:end-1], ":", 2)
		name := parts[0]
		patt := defaultPattern
		if len(parts) == 2 {
			patt = parts[1]
		}
		// Name or pattern can't be empty.
		if name == "" || patt == "" {
			return "", fmt.Errorf("mux: missing name or pattern in %q", tpl[idxs[i]:end])
		}
		// Build the regexp pattern.
		fmt.Fprintf(pattern, "%s(?P<%s>%s)", regexp.QuoteMeta(raw), varGroupName(i/2, name), patt)
	}
	// Add the remaining.
	raw := tpl[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))

	if typ == RegexpTypeQuery {
		fmt.Println(template)
		if queryVal := strings.SplitN(template, "=", 2)[1]; queryVal == "" {
			pattern.WriteString(defaultPattern)
		}
	}

	if typ != RegexpTypePrefix {
		pattern.WriteByte('$')
	}
	// Compile full regexp.
	reg, errCompile := RegexpCompileFunc(pattern.String())
	if errCompile != nil {
		return "", errCompile
	}

	// Check for capturing groups which used to work in older versions
	if reg.NumSubexp() != len(idxs)/2 {
		panic(fmt.Sprintf("route %s contains capture groups in its regexp. ", template) +
			"Only non-capturing groups are accepted: e.g. (?:pattern) instead of (pattern)")
	}
	return pattern.String(), nil
}

func braceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}

func varGroupName(idx int, name string) string {
	if name == "" {
		return "v" + strconv.Itoa(idx)
	}
	return strings.ReplaceAll(name, "-", "_")
}

func AsType[T any](src interface{}) (*T, error) {
	bs, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %T: %w", src, err)
	}
	dest := new(T)
	if err = json.Unmarshal(bs, dest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to %T: %w", dest, err)
	}
	return dest, nil
}

func FilterAnnotations(v2Annotations map[string]string) map[string]string {
	annotations := make(map[string]string)
	for key, value := range v2Annotations {
		if IsAnnotationToBeSkiped(key) {
			continue
		}
		annotations[key] = value
	}
	return annotations
}

func IsAnnotationToBeSkiped(annotation string) bool {
	for _, a := range AnnotationsTobeSkipped {
		if a == annotation {
			return true
		}
	}
	return false
}

func HasDepricatedMiddleWareOptions(m *v1alpha1.Middleware) (exists bool) {
	metaValue := reflect.ValueOf(m).Elem()
	specValue := metaValue.FieldByName("Spec")
	for name, opts := range DepricatedMiddlewareOpts {
		field := reflect.Indirect(specValue).FieldByName(name)
		if field == (reflect.Value{}) || field.IsNil() {
			continue
		}
		exists = false
		for _, opt := range opts {
			v := reflect.Indirect(field).FieldByName(opt)
			if v != (reflect.Value{}) {
				switch v.Kind() {
				case reflect.Bool:
					if value := v.Bool(); value {
						exists = true
					}
				case reflect.String:
					if value := v.String(); value != "" {
						exists = true
					}
				case reflect.Int, reflect.Int8, reflect.Int16,
					reflect.Int32, reflect.Int64:
					if v.Int() > 0 {
						exists = true
					}
				}
			}
		}
	}
	return exists
}
