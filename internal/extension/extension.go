package extension

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"text/template"
	"time"

	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/bakito/argocd-touch-extension/internal/k8s"
	sprig "github.com/go-task/slim-sprig/v3"
)

type templateConfig struct {
	name    string
	content string
}

const (
	extensionJSPath = "resources/extension-touch.js"
	fileMode        = 0o644
)

var (
	//go:embed argocd-helm-values.yaml.tpl
	tplArgocdHelmConfig string
	//go:embed argocd-server-deployment.yaml.tpl
	tplArgocdServerDeployment string
	//go:embed extension-touch.js.tpl
	tplExtension string
	//go:embed extension-proxy-rbac.yaml.tpl
	tplRBAC string

	templates = map[string]templateConfig{
		"config":     {"argocd-helm-values.yaml", tplArgocdHelmConfig},
		"deployment": {"argocd-server-deployment.yaml", tplArgocdServerDeployment},
		"extension":  {"extension-touch.js", tplExtension},
		"rbac":       {"extension-proxy-rbac.yaml", tplRBAC},
	}
)

type Error struct {
	Operation string
	Err       error
}

func (e *Error) Error() string {
	return fmt.Sprintf("extension %s failed: %v", e.Operation, e.Err)
}

type Extension interface {
	Resources() map[string]config.Resource
	ExtensionTarGz() ([]byte, string)
	ExtensionJS() []byte
	ArgoCDConfig() []byte
	ArgoCDDeployment() []byte
	ProxyRBAC() []byte
}

type extension struct {
	cfg                  config.TouchConfig
	argocdConfig         []byte
	argocdDeployment     []byte
	extensionJS          []byte
	extensionTar         []byte
	extensionTarChecksum string
	rbac                 []byte
	resourcesByGroup     map[string][]string
}

func New(cfg config.TouchConfig, cl k8s.Client, uiExtensionTemplate string) (Extension, error) {
	resources, err := cl.SetNameAndVersion(cfg.Resources)
	if err != nil {
		return nil, &Error{"version resolution", err}
	}
	cfg.Resources = resources

	ext := &extension{
		cfg:              cfg,
		resourcesByGroup: make(map[string][]string),
	}

	ext.consolidateResources()

	if err := ext.generateExtensionFiles(uiExtensionTemplate); err != nil {
		return nil, err
	}

	return ext, nil
}

func (e *extension) generateExtensionFiles(uiExtensionTemplate string) error {
	var err error

	// Generate extension JS and create tar

	uiTpl := templates["extension"]
	if uiExtensionTemplate != "" {
		data, err := os.ReadFile(uiExtensionTemplate)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		uiTpl = templateConfig{
			name:    uiExtensionTemplate,
			content: string(data),
		}
	}

	e.extensionJS, err = e.renderTemplate(uiTpl)
	if err != nil {
		return &Error{"render extension", err}
	}

	if e.extensionTar, e.extensionTarChecksum, err = e.createTar(); err != nil {
		return &Error{"create tar", err}
	}

	// Generate other template files
	if e.argocdConfig, err = e.renderTemplate(templates["config"]); err != nil {
		return &Error{"render config", err}
	}

	if e.argocdDeployment, err = e.renderTemplate(templates["deployment"]); err != nil {
		return &Error{"render deployment", err}
	}

	if e.rbac, err = e.renderTemplate(templates["rbac"]); err != nil {
		return &Error{"render rbac", err}
	}

	return nil
}

func (e *extension) renderTemplate(tpl templateConfig) ([]byte, error) {
	t, err := template.New(tpl.name).Funcs(sprig.TxtFuncMap()).Option("missingkey=error").Parse(tpl.content)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	data := map[string]any{
		"Resources":        e.cfg.Resources,
		"ServiceAddress":   e.cfg.ServiceAddress,
		"ResourcesByGroup": e.resourcesByGroup,
	}

	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *extension) ProxyRBAC() []byte {
	return e.rbac
}

func (e *extension) ArgoCDDeployment() []byte {
	return e.argocdDeployment
}

func (e *extension) ArgoCDConfig() []byte {
	return e.argocdConfig
}

func (e *extension) Resources() map[string]config.Resource {
	return e.cfg.Resources
}

func (e *extension) ExtensionJS() []byte {
	return e.extensionJS
}

func (e *extension) ExtensionTarGz() ([]byte, string) {
	return e.extensionTar, e.extensionTarChecksum
}

func (e *extension) createTar() (archive []byte, checksum string, err error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name:    extensionJSPath,
		Mode:    fileMode,
		Size:    int64(len(e.extensionJS)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, "", err
	}

	if _, err := tw.Write(e.extensionJS); err != nil {
		return nil, "", err
	}

	if err := tw.Close(); err != nil {
		return nil, "", err
	}
	if err := gw.Close(); err != nil {
		return nil, "", err
	}

	archive = buf.Bytes()
	checksum, err = calculateSHA256(archive)
	if err != nil {
		return nil, "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	return archive, checksum, nil
}

func calculateSHA256(data []byte) (string, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (e *extension) consolidateResources() {
	e.resourcesByGroup = make(map[string][]string)
	for _, resource := range e.cfg.Resources {
		sl := e.resourcesByGroup[resource.Group]
		sl = append(sl, resource.Name)
		sort.Strings(sl)
		e.resourcesByGroup[resource.Group] = sl
	}
}
