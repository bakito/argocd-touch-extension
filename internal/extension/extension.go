package extension

import (
	"archive/tar"
	"bytes"
	_ "embed"
	"text/template"
	"time"

	"github.com/bakito/argocd-touch-extension/internal/config"
)

var (
	//go:embed argocd-config.yaml.tpl
	tplArgocdConfig string
	//go:embed argocd-server-deployment.yaml.tpl
	tplArgocdServerDeployment string
	//go:embed extension-touch.js.tpl
	tplExtension string
)

func New(cfg config.TouchConfig) (Extension, error) {
	ext := &extension{cfg: cfg}

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

	return ext, nil
}

type Extension interface {
	Resources() map[string]config.Resource
	ExtensionTar() []byte
	ArgoCDConfig() []byte
	ArgoCDDeployment() []byte
}
type extension struct {
	cfg              config.TouchConfig
	argocdConfig     []byte
	argocdDeployment []byte
	extensionTar     []byte
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
