package converter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/databotic/traefik-migration-tool/internal/utils"
	containous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var rewriteHTTPFunc = map[string]func(v []string) string{
	"PathPrefix":    rewritePathPrefix,
	"Path":          rewritePath,
	"Host":          rewriteHost,
	"HostRegexp":    rewriteHost,
	"Query":         rewriteQuery,
	"Method":        rewriteMethod,
	"Headers":       rewriteHeaders,
	"HeadersRegexp": rewriteHeadersRegexp,
	"ClientIP":      rewriteClientIP,
}

type IngressRoute struct {
	muxer *httpmuxer.Muxer
}

func NewIngressRoute() (*IngressRoute, error) {
	muxer, err := httpmuxer.NewMuxer()
	if err != nil {
		return nil, err
	}

	return &IngressRoute{
		muxer: muxer,
	}, nil
}

func (t *IngressRoute) Transform(object runtime.Object) (runtime.Object, error) {
	ingressRoute, ok := object.(*containous.IngressRoute)
	if !ok {
		return nil, fmt.Errorf("err")
	}

	v3IngressRoute := &traefikio.IngressRoute{
		TypeMeta: v1.TypeMeta{Kind: ingressRoute.Kind, APIVersion: utils.APIVersion},
		ObjectMeta: v1.ObjectMeta{
			Name: ingressRoute.ObjectMeta.Name, Namespace: ingressRoute.ObjectMeta.Namespace,
			Annotations: utils.FilterAnnotations(ingressRoute.Annotations), Labels: ingressRoute.Labels,
		},
		Spec: traefikio.IngressRouteSpec{
			EntryPoints: ingressRoute.Spec.EntryPoints,
		},
	}

	tls, err := utils.AsType[traefikio.TLS](ingressRoute.Spec.TLS)
	if err != nil {
		return nil, err
	}

	// to address empty fild when writing as yml
	if len(tls.Domains) > 0 || tls.SecretName != "" || tls.Store != nil {
		v3IngressRoute.Spec.TLS = tls
	}

	routes, err := t.transformRoute(ingressRoute.Spec.Routes)
	if err != nil {
		return nil, err
	}

	v3IngressRoute.Spec.Routes = routes
	return v3IngressRoute, nil
}

func (t *IngressRoute) transformRoute(v2Routes []containous.Route) ([]traefikio.Route, error) {
	var routes []traefikio.Route
	for _, r := range v2Routes {
		if err := t.checkRoute(r.Match, "v2"); err != nil {
			return nil, err
		}

		route, err := utils.AsType[traefikio.Route](r)
		if err != nil {
			return nil, err
		}

		if m := t.transformRule(r.Match); len(m) > 0 {
			if err := t.checkRoute(m, "v3"); err != nil {
				return nil, err
			}
			route.Match = m
		}

		routes = append(routes, *route)
	}
	return routes, nil
}

func (t *IngressRoute) transformRule(rule string) string {
	return rulePattern.ReplaceAllStringFunc(rule, func(match string) string {
		functionName := rulePattern.FindStringSubmatch(match)[1]
		vaules := rulePattern.FindStringSubmatch(match)[2]
		arguments := strings.Split(strings.ReplaceAll(strings.ReplaceAll(vaules, "`", ""), " ", ""), ",")

		if funcName, ok := rewriteHTTPFunc[functionName]; ok {
			return funcName(arguments)
		}

		return match
	})
}

func rewriteHeaders(values []string) string {
	return fmt.Sprintf("Header(`%s`, `%s`)", values[0], values[1])
}

func rewriteHeadersRegexp(values []string) string {
	return fmt.Sprintf("HeaderRegexp(`%s`, `%s`)", values[0], values[1])
}

func rewritePathPrefix(values []string) string {
	var transformedArgs []string
	for _, v := range values {
		if regexPattern.MatchString(v) {
			pattern, err := utils.RouteRegexp(v, utils.RegexpTypePrefix)
			if err != nil {
				panic(err)
			}
			transformedArg := fmt.Sprintf("PathRegexp(`%s`)", pattern)
			transformedArgs = append(transformedArgs, transformedArg)
		} else {
			transformedArg := fmt.Sprintf("PathPrefix(`%s`)", v)
			transformedArgs = append(transformedArgs, transformedArg)
		}
	}

	if len(transformedArgs) > 1 {
		return "(" + strings.Join(transformedArgs, " || ") + ")"
	}

	return strings.Join(transformedArgs, " ")
}

func rewritePath(values []string) string {
	var transformedArgs []string

	for _, v := range values {
		if regexPattern.MatchString(v) {
			pattern, err := utils.RouteRegexp(v, utils.RegexpTypePath)
			if err != nil {
				panic(err)
			}

			transformedArg := fmt.Sprintf("PathRegexp(`%s`)", pattern)
			transformedArgs = append(transformedArgs, transformedArg)
		} else {
			transformedArg := fmt.Sprintf("Path(`%s`)", v)
			transformedArgs = append(transformedArgs, transformedArg)
		}
	}

	if len(transformedArgs) > 1 {
		return "(" + strings.Join(transformedArgs, " || ") + ")"
	}

	return strings.Join(transformedArgs, " ")
}

func rewriteMethod(values []string) string {
	var transformedArgs []string
	for _, v := range values {
		transformedArg := fmt.Sprintf("Method(`%s`)", v)
		transformedArgs = append(transformedArgs, transformedArg)
	}

	if len(transformedArgs) > 1 {
		return "(" + strings.Join(transformedArgs, " || ") + ")"
	}

	return strings.Join(transformedArgs, " ")
}

func rewriteHost(values []string) string {
	var transformedArgs []string
	for _, v := range values {
		if regexPattern.MatchString(v) {
			pattern, err := utils.RouteRegexp(v, utils.RegexpTypeHost)
			if err != nil {
				panic(err)
			}
			transformedArg := fmt.Sprintf("HostRegexp(`%s`)", pattern)
			transformedArgs = append(transformedArgs, transformedArg)
		} else {
			transformedArg := fmt.Sprintf("Host(`%s`)", v)
			transformedArgs = append(transformedArgs, transformedArg)
		}
	}

	if len(transformedArgs) > 1 {
		return "(" + strings.Join(transformedArgs, " || ") + ")"
	}

	return strings.Join(transformedArgs, " ")
}

func rewriteClientIP(values []string) string {
	var transformedArgs []string
	for _, v := range values {
		transformedArg := fmt.Sprintf("ClientIP(`%s`)", v)
		transformedArgs = append(transformedArgs, transformedArg)
	}

	if len(transformedArgs) > 1 {
		return "(" + strings.Join(transformedArgs, " || ") + ")"
	}

	return strings.Join(transformedArgs, " ")
}

func rewriteQuery(values []string) string {
	var transformedArgs []string

	for _, value := range values {
		key, queryValue := splitQueryKeyValue(value)
		transformedArg := formatQuery(key, queryValue)
		transformedArgs = append(transformedArgs, transformedArg)
	}

	if len(transformedArgs) > 1 {
		return "(" + strings.Join(transformedArgs, " || ") + ")"
	}
	return strings.Join(transformedArgs, " ")
}

func splitQueryKeyValue(query string) (key, value string) {
	parts := strings.SplitN(query, "=", 2)
	key = parts[0]
	if len(parts) > 1 {
		value = parts[1]
	}
	return
}

func formatQuery(key, value string) string {
	var formatted string
	if value != "" {
		pattern, err := utils.RouteRegexp(value, utils.RegexpTypeQuery)
		if err != nil {
			panic(err)
		}

		if regexPattern.MatchString(value) {
			formatted = fmt.Sprintf("QueryRegexp(`%s`, `%s`)", key, pattern)
		} else {
			formatted = fmt.Sprintf("Query(`%s`, `%s`)", key, value)
		}
	} else {
		formatted = fmt.Sprintf("Query(`%s`)", key)
	}
	return formatted
}

func (t *IngressRoute) checkRoute(rule string, syntax string) error {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	if err := t.muxer.AddRoute(rule, syntax, 0, handler); err != nil {
		return err
	}

	return nil
}
