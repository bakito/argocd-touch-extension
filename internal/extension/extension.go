package extension

import (
	"archive/tar"
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
	"time"

	"github.com/bakito/argocd-touch-extension/internal/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
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
)

func New(cfg config.TouchConfig, dcl *discovery.DiscoveryClient) (Extension, error) {
	for key, res := range cfg.Resources {
		v, name, err := getServerPreferredVersion(dcl, res.Group, res.Kind)
		if err != nil {
			return nil, err
		}
		if res.Version == "" {
			res.Version = v
		}
		res.Name = name
		cfg.Resources[key] = res
	}

	ext := &extension{cfg: cfg}
	ext.consolidate()

	extJS, err := ext.render("extension-touch.js", tplExtension)
	if err != nil {
		return nil, err
	}
	ext.extensionTar, err = ext.createTar(extJS)
	if err != nil {
		return nil, err
	}

	ext.argocdConfig, err = ext.render("argocd-config.yaml", tplArgocdConfig)
	if err != nil {
		return nil, err
	}
	ext.argocdDeployment, err = ext.render("argocd-server-deployment.yaml", tplArgocdServerDeployment)
	if err != nil {
		return nil, err
	}
	// FIXME consolidate by group for rbac
	ext.rbac, err = ext.render("extension-proxy-rbac.yaml", tplRBAC)
	if err != nil {
		return nil, err
	}

	return ext, nil
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
	ResourcesByGroup map[string][]string
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

func (e *extension) render(name, templ string) ([]byte, error) {
	t, err := template.New(name).Parse(templ)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer

	if err := t.Execute(&tpl, e.cfg); err != nil {
		return nil, err
	}
	return tpl.Bytes(), nil
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
		Name:    "resources/extension-touch.js",
		Mode:    0o644,
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

func (e *extension) consolidate() {
	e.ResourcesByGroup = make(map[string][]string)
	for _, resource := range e.cfg.Resources {
		e.ResourcesByGroup[resource.Group] = append(e.ResourcesByGroup[resource.Group], resource.Name)
	}
}

func getServerPreferredVersion(
	discoveryClient *discovery.DiscoveryClient,
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
