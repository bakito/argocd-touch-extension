package extension

import (
	"archive/tar"
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"text/template"
	"time"

	"github.com/bakito/argocd-touch-extension/internal/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
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
	//go:embed argocd-config.yaml.tpl
	tplArgocdConfig string
	//go:embed argocd-server-deployment.yaml.tpl
	tplArgocdServerDeployment string
	//go:embed extension-touch.js.tpl
	tplExtension string
	//go:embed extension-proxy-rbac.yaml.tpl
	tplRBAC string

	templates = map[string]templateConfig{
		"config":     {"argocd-config.yaml", tplArgocdConfig},
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
	ExtensionTar() []byte
	ArgoCDConfig() []byte
	ArgoCDDeployment() []byte
	ProxyRBAC() []byte
}
type extension struct {
	cfg              config.TouchConfig
	argocdConfig     []byte
	argocdDeployment []byte
	extensionTar     []byte
	rbac             []byte
	resourcesByGroup map[string][]string
}

func New(cfg config.TouchConfig, dcl discovery.DiscoveryInterface) (Extension, error) {
	ext := &extension{
		cfg:              cfg,
		resourcesByGroup: make(map[string][]string),
	}

	if err := ext.resolveResourceVersions(dcl); err != nil {
		return nil, &Error{"version resolution", err}
	}

	ext.consolidateResources()

	if err := ext.generateExtensionFiles(); err != nil {
		return nil, err
	}

	return ext, nil
}

func (e *extension) resolveResourceVersions(dcl discovery.DiscoveryInterface) error {
	for key, res := range e.cfg.Resources {
		version, name, err := getServerPreferredVersion(dcl, res.Group, res.Kind)
		if err != nil {
			return err
		}
		if res.Version == "" {
			res.Version = version
		}
		res.Name = name
		e.cfg.Resources[key] = res
	}
	return nil
}

func (e *extension) generateExtensionFiles() error {
	var err error

	// Generate extension JS and create tar
	extJS, err := e.renderTemplate(templates["extension"])
	if err != nil {
		return &Error{"render extension", err}
	}

	if e.extensionTar, err = e.createTar(extJS); err != nil {
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
	t, err := template.New(tpl.name).Parse(tpl.content)
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

func (e *extension) ExtensionTar() []byte {
	return e.extensionTar
}

func (e *extension) createTar(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	hdr := &tar.Header{
		Name:    extensionJSPath,
		Mode:    fileMode,
		Size:    int64(len(data)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}

	if _, err := tw.Write(data); err != nil {
		return nil, err
	}

	_ = tw.Close()
	return buf.Bytes(), nil
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

func getServerPreferredVersion(
	discoveryClient discovery.DiscoveryInterface,
	group, kind string,
) (version, name string, err error) {
	resources, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return "", "", fmt.Errorf("failed to get server preferred resources: %w", err)
	}

	for _, list := range resources {
		if list == nil {
			continue
		}

		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		if gv.Group == group {
			for _, r := range list.APIResources {
				if r.Kind == kind {
					return gv.Version, r.Name, nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("no preferred version found for group %s and kind %s", group, kind)
}
