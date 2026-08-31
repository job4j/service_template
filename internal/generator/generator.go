package generator

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const templateRoot = "templates/project"

var serviceNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

//go:embed templates/project/** templates/project/.gitignore.tmpl
var projectTemplates embed.FS

type Config struct {
	Name       string
	ModulePath string
	OutputDir  string
	HTTPPort   int
}

type templateData struct {
	ServiceName string
	ModulePath  string
	HTTPPort    int
}

func Generate(config Config) error {
	if config.HTTPPort == 0 {
		config.HTTPPort = 8080
	}
	if err := validateConfig(config); err != nil {
		return err
	}

	destination := config.OutputDir
	if destination == "" {
		destination = config.Name
	}

	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}

	exists, err := prepareDestination(absDestination)
	if err != nil {
		return err
	}

	data := templateData{
		ServiceName: config.Name,
		ModulePath:  config.ModulePath,
		HTTPPort:    config.HTTPPort,
	}

	if exists {
		if err := renderProject(absDestination, data); err != nil {
			_ = removeGeneratedContents(absDestination)
			return err
		}
		return nil
	}

	parent := filepath.Dir(absDestination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	temporary, err := os.MkdirTemp(parent, ".servicegen-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(temporary)

	if err := renderProject(temporary, data); err != nil {
		return err
	}
	if err := os.Rename(temporary, absDestination); err != nil {
		return fmt.Errorf("move generated service to %s: %w", absDestination, err)
	}
	return nil
}

func validateConfig(config Config) error {
	if !serviceNamePattern.MatchString(config.Name) {
		return fmt.Errorf("invalid service name %q: use lowercase letters, digits, and hyphens", config.Name)
	}
	if err := validateModulePath(config.ModulePath); err != nil {
		return err
	}
	if config.HTTPPort < 1 || config.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port %d: expected a value from 1 to 65535", config.HTTPPort)
	}
	return nil
}

func validateModulePath(modulePath string) error {
	if modulePath == "" {
		return errors.New("module path is required")
	}
	if strings.ContainsAny(modulePath, "\\ \t\r\n") {
		return fmt.Errorf("invalid module path %q", modulePath)
	}
	if !strings.Contains(modulePath, "/") || strings.HasPrefix(modulePath, "/") || strings.HasSuffix(modulePath, "/") {
		return fmt.Errorf("invalid module path %q: expected a full path such as github.com/user/service", modulePath)
	}
	for _, part := range strings.Split(modulePath, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("invalid module path %q", modulePath)
		}
	}
	return nil
}

func prepareDestination(destination string) (bool, error) {
	info, err := os.Stat(destination)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect output directory: %w", err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("output path %s already exists and is not a directory", destination)
	}

	entries, err := os.ReadDir(destination)
	if err != nil {
		return false, fmt.Errorf("read output directory: %w", err)
	}
	if len(entries) != 0 {
		return false, fmt.Errorf("output directory %s is not empty", destination)
	}
	return true, nil
}

func renderProject(destination string, data templateData) error {
	return fs.WalkDir(projectTemplates, templateRoot, func(templatePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(templateRoot, templatePath)
		if err != nil {
			return fmt.Errorf("resolve template path %s: %w", templatePath, err)
		}
		relativePath = strings.TrimSuffix(relativePath, ".tmpl")
		outputPath := filepath.Join(destination, relativePath)

		content, err := projectTemplates.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", templatePath, err)
		}
		parsed, err := template.New(templatePath).Option("missingkey=error").Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", templatePath, err)
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return fmt.Errorf("create directory for %s: %w", outputPath, err)
		}
		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("create %s: %w", outputPath, err)
		}
		executeErr := parsed.Execute(file, data)
		closeErr := file.Close()
		if executeErr != nil {
			return fmt.Errorf("render %s: %w", outputPath, executeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", outputPath, closeErr)
		}
		return nil
	})
}

func removeGeneratedContents(destination string) error {
	entries, err := os.ReadDir(destination)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(destination, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
