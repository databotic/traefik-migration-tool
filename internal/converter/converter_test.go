package converter

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/databotic/traefik-migration-tool/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateExpected = flag.Bool("update_expected", false, "Update expected files in testdata")

type TestStruct struct {
	ingressRouteFile string
}

func testFile(c TestStruct, t *testing.T) {
	outputDir := filepath.Join("fixtures", "output")
	inputFile := filepath.Join("fixtures", "input", c.ingressRouteFile)

	stat, err := os.Stat(inputFile)
	require.NoError(t, err)

	file, err := os.OpenFile(inputFile, os.O_RDONLY, stat.Mode())
	require.NoError(t, err)

	objects, err := parser.ParseManifest(file)
	require.NoError(t, err)

	converter, err := New()
	require.NoError(t, err)

	converted, err := converter.Do(objects)
	require.NoError(t, err)

	var fragments []string
	for _, object := range converted {
		s, err := converter.EncodeYaml(object)
		require.NoError(t, err)

		fragments = append(fragments, string(s))
	}

	fixtureFile := filepath.Join(outputDir, c.ingressRouteFile)
	data := strings.Join(fragments, "---\n")

	if *updateExpected {
		require.NoError(t, os.WriteFile(fixtureFile, []byte(data), 0o666))
	}

	yml, err := os.ReadFile(fixtureFile)
	require.NoError(t, err)
	assert.YAMLEq(t, string(yml), string(data))
}

func TestIngressRoutes(t *testing.T) {
	testCases := []TestStruct{
		{
			ingressRouteFile: "ingressroute_host.yaml",
		},
		{
			ingressRouteFile: "ingressroute_path.yaml",
		},
		{
			ingressRouteFile: "ingressroute_pathprefix.yaml",
		},
		{
			ingressRouteFile: "ingressroute_methods.yaml",
		},
		{
			ingressRouteFile: "ingressroute_multi_doc.yaml",
		},
		{
			ingressRouteFile: "ingressroute_path_with_tls.yaml",
		},
	}
	for _, test := range testCases {
		t.Run(test.ingressRouteFile, func(t *testing.T) {
			testFile(test, t)
		})
	}
}

func TestMiddleWares(t *testing.T) {
	testCases := []TestStruct{
		{
			ingressRouteFile: "middleware_strip_prefix.yaml",
		},
		{
			ingressRouteFile: "middleware_raqlimit.yaml",
		},
		{
			ingressRouteFile: "middleware_ipwhitelist.yaml",
		},
	}
	for _, test := range testCases {
		t.Run(test.ingressRouteFile, func(t *testing.T) {
			testFile(test, t)
		})
	}
}
