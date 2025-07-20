package extension

import (
	"bytes"
	"os"
	"testing"
	"text/template"

	"github.com/bakito/argocd-touch-extension/internal/config"
	sprig "github.com/go-task/slim-sprig/v3"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateCfg templateConfig
		config      config.TouchConfig
		groupedRes  map[string][]string
		expectErr   bool
		expected    string
	}{
		{
			name: "valid template",
			templateCfg: templateConfig{
				name:    "validTpl",
				content: "Hello, {{.ServiceAddress}}!",
			},
			config: config.TouchConfig{
				ServiceAddress: "test-service",
				Resources:      map[string]config.Resource{},
			},
			groupedRes: map[string][]string{},
			expectErr:  false,
			expected:   "Hello, test-service!",
		},
		{
			name: "template with resources",
			templateCfg: templateConfig{
				name:    "resourceTpl",
				content: "Resources: {{range $key, $value := .Resources}}{{$key}} {{$value.Name}} {{end}}",
			},
			config: config.TouchConfig{
				Resources: map[string]config.Resource{
					"res1": {Group: "group1", Name: "resource1"},
					"res2": {Group: "group2", Name: "resource2"},
				},
			},
			groupedRes: map[string][]string{
				"group1": {"resource1"},
				"group2": {"resource2"},
			},
			expectErr: false,
			expected:  "Resources: res1 resource1 res2 resource2 ",
		},
		{
			name: "empty template",
			templateCfg: templateConfig{
				name:    "emptyTpl",
				content: "",
			},
			config: config.TouchConfig{
				ServiceAddress: "test-service",
				Resources:      map[string]config.Resource{},
			},
			groupedRes: map[string][]string{},
			expectErr:  false,
			expected:   "",
		},
		{
			name: "template with invalid syntax",
			templateCfg: templateConfig{
				name:    "invalidSyntax",
				content: "{{.ServiceAddress",
			},
			config: config.TouchConfig{
				ServiceAddress: "test-service",
				Resources:      map[string]config.Resource{},
			},
			groupedRes: map[string][]string{},
			expectErr:  true,
			expected:   "",
		},
		{
			name: "template with missing data",
			templateCfg: templateConfig{
				name:    "missingData",
				content: "Address: {{.NonExistent}}!",
			},
			config:     config.TouchConfig{},
			groupedRes: map[string][]string{},
			expectErr:  true,
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &extension{
				cfg:              tt.config,
				resourcesByGroup: tt.groupedRes,
			}
			result, err := e.renderTemplate(tt.templateCfg)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestTemplateExecutionWithSprig(t *testing.T) {
	tpl, err := template.New("sprigTpl").Funcs(sprig.TxtFuncMap()).Parse("{{.Name | upper}}")
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	data := map[string]any{
		"Name": "example",
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}
	expected := "EXAMPLE"
	if buf.String() != expected {
		t.Fatalf("expected %q, got %q", expected, buf.String())
	}
}

func TestConsolidateResources(t *testing.T) {
	tests := []struct {
		name     string
		input    config.TouchConfig
		expected map[string][]string
	}{
		{
			name: "single resource in one group",
			input: config.TouchConfig{
				Resources: map[string]config.Resource{
					"res1": {Group: "group1", Name: "resource1"},
				},
			},
			expected: map[string][]string{
				"group1": {"resource1"},
			},
		},
		{
			name: "multiple resources in one group",
			input: config.TouchConfig{
				Resources: map[string]config.Resource{
					"res1": {Group: "group1", Name: "resource1"},
					"res2": {Group: "group1", Name: "resource2"},
				},
			},
			expected: map[string][]string{
				"group1": {"resource1", "resource2"},
			},
		},
		{
			name: "resources across multiple groups",
			input: config.TouchConfig{
				Resources: map[string]config.Resource{
					"res1": {Group: "group1", Name: "resource1"},
					"res2": {Group: "group2", Name: "resource2"},
				},
			},
			expected: map[string][]string{
				"group1": {"resource1"},
				"group2": {"resource2"},
			},
		},
		{
			name: "multiple resources in multiple groups",
			input: config.TouchConfig{
				Resources: map[string]config.Resource{
					"res1": {Group: "group1", Name: "resource1"},
					"res2": {Group: "group2", Name: "resource2"},
					"res3": {Group: "group1", Name: "resource3"},
				},
			},
			expected: map[string][]string{
				"group1": {"resource1", "resource3"},
				"group2": {"resource2"},
			},
		},
		{
			name: "no resources",
			input: config.TouchConfig{
				Resources: map[string]config.Resource{},
			},
			expected: map[string][]string{},
		},
		{
			name: "unsorted resource names",
			input: config.TouchConfig{
				Resources: map[string]config.Resource{
					"res2": {Group: "group1", Name: "resource2"},
					"res1": {Group: "group1", Name: "resource1"},
				},
			},
			expected: map[string][]string{
				"group1": {"resource1", "resource2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &extension{
				cfg: tt.input,
			}
			e.consolidateResources()
			if len(e.resourcesByGroup) != len(tt.expected) {
				t.Fatalf("expected %d groups, got %d", len(tt.expected), len(e.resourcesByGroup))
			}
			for group, names := range tt.expected {
				if !equalSlices(e.resourcesByGroup[group], names) {
					t.Errorf("for group %q, expected %v, got %v", group, names, e.resourcesByGroup[group])
				}
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGenerateExtensionFiles(t *testing.T) {
	tests := []struct {
		name         string
		config       config.TouchConfig
		templateFile string
		templates    map[string]templateConfig
		expectErr    bool
	}{
		{
			name: "valid ui extension template",
			config: config.TouchConfig{
				ServiceAddress: "valid-service",
				Resources: map[string]config.Resource{
					"res1": {Group: "group1", Name: "resource1"},
				},
			},
			templateFile: "",
			templates: map[string]templateConfig{
				"extension":  {name: "extensionTpl", content: "Service: {{.ServiceAddress}}"},
				"config":     {name: "configTpl", content: "Config for {{.ServiceAddress}}"},
				"deployment": {name: "deploymentTpl", content: "Deploy {{.ResourcesByGroup.group1}}"},
				"rbac":       {name: "rbacTpl", content: "RBAC for {{.ServiceAddress}}"},
			},
			expectErr: false,
		},
		{
			name: "missing key in template",
			config: config.TouchConfig{
				ServiceAddress: "invalid-service",
				Resources:      map[string]config.Resource{},
			},
			templateFile: "",
			templates: map[string]templateConfig{
				"extension": {name: "missingKeyTpl", content: "Service: {{.NonExistentKey}}"},
			},
			expectErr: true,
		},
		{
			name: "invalid template syntax",
			config: config.TouchConfig{
				ServiceAddress: "test-service",
				Resources:      map[string]config.Resource{},
			},
			templateFile: "",
			templates: map[string]templateConfig{
				"extension": {name: "invalidSyntaxTpl", content: "Hello {{.ServiceAddress"},
			},
			expectErr: true,
		},
		{
			name: "valid template from file",
			config: config.TouchConfig{
				ServiceAddress: "file-service",
				Resources: map[string]config.Resource{
					"res1": {Group: "group1", Name: "resource1"},
				},
			},
			templateFile: "test_template.tpl",
			templates: map[string]templateConfig{
				"extension": {name: "ignoredTpl", content: "This will be replaced."},
			},
			expectErr: false,
		},
		{
			name: "file read failure",
			config: config.TouchConfig{
				ServiceAddress: "file-error-service",
				Resources:      map[string]config.Resource{},
			},
			templateFile: "nonexistent_file.tpl",
			templates:    nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates = tt.templates

			e := &extension{
				cfg: tt.config,
			}

			if tt.templateFile == "test_template.tpl" {
				testContent := "File-based template: {{.ServiceAddress}}"
				err := os.WriteFile(tt.templateFile, []byte(testContent), os.ModePerm)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tt.templateFile)
			}

			e.consolidateResources()
			err := e.generateExtensionFiles(tt.templateFile)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if tt.templates != nil && tt.templates["extension"].content == "Service: {{.ServiceAddress}}" {
				expectedContent := "Service: " + tt.config.ServiceAddress
				if string(e.extensionJS) != expectedContent {
					t.Errorf("expected extensionJS content %q, got %q", expectedContent, string(e.extensionJS))
				}
			}
		})
	}
}
