package server

import (
	"net/http"

	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/gin-gonic/gin"
)

const (
	contentTypeYAML = "application/yaml"
	contentTypeTAR  = "application/x-tar"

	extensionFileName = "extension.tar.gz"
	extensionChecksum = "extension_checksum.txt"
)

// createHandler creates a gin.HandlerFunc that returns data with a specified content type.
func createHandler(contentType string, dataFn func() []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, contentType, dataFn())
	}
}

// configHandler returns extension configuration.
func configHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeYAML, ext.ArgoCDConfig)
}

// deploymentHandler returns extension deployment configuration.
func deploymentHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeYAML, ext.ArgoCDDeployment)
}

// tarHandler returns extension tar archive.
func tarHandler(ext extension.Extension) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Disposition", "attachment; filename="+extensionFileName)
		tarGZ, _ := ext.ExtensionTarGz()
		c.Data(http.StatusOK, contentTypeTAR, tarGZ)
	}
}

// tarChecksumHandler returns extension tar archive.
func tarChecksumHandler(ext extension.Extension) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Disposition", "attachment; filename="+extensionChecksum)
		_, cs := ext.ExtensionTarGz()
		c.String(http.StatusOK, "%s  %s", cs, extensionChecksum)
	}
}

// rbacHandler returns extension RBAC configuration.
func rbacHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeYAML, ext.ProxyRBAC)
}
