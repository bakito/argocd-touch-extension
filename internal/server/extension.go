package server

import (
	"net/http"

	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/gin-gonic/gin"
)

const (
	contentTypeYAML = "application/yaml"
	contentTypeTAR  = "application/x-tar"
)

// createHandler creates a gin.HandlerFunc that returns data with specified content type.
func createHandler(contentType string, dataFn func() []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, contentType, dataFn())
	}
}

// ConfigHandler returns extension configuration.
func configHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeYAML, ext.ArgoCDConfig)
}

// DeploymentHandler returns extension deployment configuration.
func deploymentHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeYAML, ext.ArgoCDDeployment)
}

// TarHandler returns extension tar archive.
func tarHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeTAR, ext.ExtensionTar)
}

// RBACHandler returns extension RBAC configuration.
func rbacHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeYAML, ext.ProxyRBAC)
}
