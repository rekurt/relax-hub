package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type documentedRoute struct {
	method string
	path   string
}

func TestSwaggerDocumentsAllAPIV1Routes(t *testing.T) {
	repoRoot := testRepoRoot(t)
	routerRoutes := extractRouterRoutes(t, repoRoot)
	swaggerRoutes := extractSwaggerRoutes(t, repoRoot)

	var prefixedPaths []string
	for route := range swaggerRoutes {
		if strings.HasPrefix(route.path, "/api/v1/") {
			prefixedPaths = append(prefixedPaths, fmt.Sprintf("%s %s", strings.ToUpper(route.method), route.path))
		}
	}
	sort.Strings(prefixedPaths)
	if len(prefixedPaths) > 0 {
		t.Fatalf("swagger paths must not include /api/v1 prefix because @BasePath already sets it:\n%s", strings.Join(prefixedPaths, "\n"))
	}

	var missing []string
	for route := range routerRoutes {
		if _, ok := swaggerRoutes[route]; !ok {
			missing = append(missing, fmt.Sprintf("%s %s", strings.ToUpper(route.method), route.path))
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("API routes missing from swagger:\n%s", strings.Join(missing, "\n"))
	}
}

func testRepoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func extractRouterRoutes(t *testing.T, repoRoot string) map[documentedRoute]struct{} {
	t.Helper()

	files := []string{
		filepath.Join(repoRoot, "internal/server/routes_api.go"),
		filepath.Join(repoRoot, "internal/server/routes_my.go"),
		filepath.Join(repoRoot, "internal/server/routes_admin.go"),
	}
	methodNames := map[string]string{
		"Get":    "get",
		"Post":   "post",
		"Put":    "put",
		"Patch":  "patch",
		"Delete": "delete",
	}
	routePrefixRe := regexp.MustCompile(`\.Route\("([^"]+)"`)
	methodRe := regexp.MustCompile(`\.(Get|Post|Put|Patch|Delete)\("([^"]+)"`)

	routes := make(map[documentedRoute]struct{})
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}

		prefix := ""
		for _, line := range strings.Split(string(body), "\n") {
			if matches := routePrefixRe.FindStringSubmatch(line); len(matches) == 2 {
				prefix = strings.TrimRight(matches[1], "/")
			}
			matches := methodRe.FindStringSubmatch(line)
			if len(matches) != 3 {
				continue
			}
			routes[documentedRoute{
				method: methodNames[matches[1]],
				path:   prefix + matches[2],
			}] = struct{}{}
		}
	}
	return routes
}

func extractSwaggerRoutes(t *testing.T, repoRoot string) map[documentedRoute]struct{} {
	t.Helper()

	body, err := os.ReadFile(filepath.Join(repoRoot, "docs/swagger.json"))
	if err != nil {
		t.Fatalf("read swagger.json: %v", err)
	}

	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(body, &spec); err != nil {
		t.Fatalf("parse swagger.json: %v", err)
	}

	httpMethods := map[string]struct{}{
		"get":    {},
		"post":   {},
		"put":    {},
		"patch":  {},
		"delete": {},
	}
	routes := make(map[documentedRoute]struct{})
	for path, operations := range spec.Paths {
		for method := range operations {
			if _, ok := httpMethods[method]; !ok {
				continue
			}
			routes[documentedRoute{method: method, path: path}] = struct{}{}
		}
	}
	return routes
}
