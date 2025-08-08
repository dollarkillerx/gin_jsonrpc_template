package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Data struct {
	Module   string
	AppName  string
	HTTPPort string
}

// Generate creates a new project at outDir using provided data
func Generate(outDir string, data Data) error {
	if data.Module == "" {
		return fmt.Errorf("module is required")
	}
	if data.AppName == "" {
		data.AppName = "app"
	}
	if data.HTTPPort == "" {
		data.HTTPPort = "8080"
	}

	for path, tmpl := range templates {
		fullPath := filepath.Join(outDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", filepath.Dir(fullPath), err)
		}

		// Render template
		content, err := render(tmpl, data)
		if err != nil {
			return fmt.Errorf("render %s: %w", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write file %s: %w", fullPath, err)
		}
	}
	return nil
}

func render(tmpl string, data Data) (string, error) {
	t, err := template.New("").Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	}).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
