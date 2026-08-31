package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCreatesService(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "sharetrip-contract")
	config := Config{
		Name:       "sharetrip-contract",
		ModulePath: "github.com/student/sharetrip-contract",
		OutputDir:  destination,
	}

	if err := Generate(config); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	assertFileContains(t, filepath.Join(destination, "go.mod"), "module github.com/student/sharetrip-contract")
	assertFileContains(t, filepath.Join(destination, "cmd", "main.go"), `contractapi "github.com/student/sharetrip-contract/api"`)
	assertFileContains(t, filepath.Join(destination, "api", "contract.yaml"), "title: sharetrip-contract API")
	assertFileContains(t, filepath.Join(destination, "Makefile"), "OAPI_CODEGEN_VERSION := v2.8.0")
	assertFileContains(t, filepath.Join(destination, ".gitignore"), "/bin/")
	assertFileContains(t, filepath.Join(destination, ".env.example"), "HTTP_PORT=8080")
	assertFileContains(t, filepath.Join(destination, "internal", "middleware", "request_context.go"), "X-Request-ID")
	assertFileContains(t, filepath.Join(destination, "observability", "prometheus-target.yml"), "host.docker.internal:8080")

	requiredFiles := []string{
		"Dockerfile",
		filepath.Join("api", "health.go"),
		filepath.Join("api", "openapi.codegen.yaml"),
		filepath.Join("internal", "service", "doc.go"),
		filepath.Join("internal", "domain", "doc.go"),
		filepath.Join("internal", "repository", "doc.go"),
		filepath.Join("internal", "repository", "entity", "doc.go"),
		filepath.Join("migrations", "README.md"),
		filepath.Join("internal", "config", "config.go"),
		filepath.Join("internal", "observability", "logging", "logger.go"),
		filepath.Join("internal", "observability", "tracing", "tracing.go"),
		filepath.Join("internal", "observability", "metrics", "metrics.go"),
		filepath.Join("observability", "grafana-dashboard.json"),
	}
	for _, relativePath := range requiredFiles {
		if _, err := os.Stat(filepath.Join(destination, relativePath)); err != nil {
			t.Errorf("generated file %s: %v", relativePath, err)
		}
	}
}

func TestGenerateUsesEmptyDirectory(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "empty")
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}

	err := Generate(Config{
		Name:       "billing-service",
		ModulePath: "github.com/student/billing-service",
		OutputDir:  destination,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	assertFileContains(t, filepath.Join(destination, "go.mod"), "github.com/student/billing-service")
}

func TestGenerateRefusesNonEmptyDirectory(t *testing.T) {
	destination := t.TempDir()
	existing := filepath.Join(destination, "keep.txt")
	if err := os.WriteFile(existing, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Generate(Config{
		Name:       "billing-service",
		ModulePath: "github.com/student/billing-service",
		OutputDir:  destination,
	})
	if err == nil || !strings.Contains(err.Error(), "is not empty") {
		t.Fatalf("Generate() error = %v, want non-empty directory error", err)
	}
	assertFileContains(t, existing, "keep")
}

func TestGenerateValidatesConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "uppercase service name",
			config: Config{
				Name:       "BillingService",
				ModulePath: "github.com/student/billing-service",
			},
		},
		{
			name: "short module path",
			config: Config{
				Name:       "billing-service",
				ModulePath: "billing-service",
			},
		},
		{
			name: "module path with spaces",
			config: Config{
				Name:       "billing-service",
				ModulePath: "github.com/student/billing service",
			},
		},
		{
			name:   "invalid HTTP port",
			config: Config{Name: "billing-service", ModulePath: "github.com/student/billing-service", HTTPPort: 70000},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.config.OutputDir = filepath.Join(t.TempDir(), "output")
			if err := Generate(test.config); err == nil {
				t.Fatal("Generate() error = nil, want validation error")
			}
		})
	}
}

func assertFileContains(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), expected) {
		t.Errorf("%s does not contain %q", path, expected)
	}
}
