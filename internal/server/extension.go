package server

import (
	"net/http"

	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/gin-gonic/gin"
)

const (
	contentTypeYAML = "application/yaml"
	contentTypeJS   = "application/javascript"
	contentTypeTAR  = "application/x-tar"

	extensionFileName = "extension.tar.gz"
	ExtensionChecksum = "extension_checksum.txt"
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

// jsHandler returns extension js extension.
func jsHandler(ext extension.Extension) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Disposition", "attachment; filename="+extension.ExtensionJS)
		js, _ := ext.ExtensionJS()
		c.Data(http.StatusOK, contentTypeJS, js)
	}
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
		c.Header("Content-Disposition", "attachment; filename="+ExtensionChecksum)
		_, csTar := ext.ExtensionTarGz()
		_, csjs := ext.ExtensionJS()
		c.String(http.StatusOK, "%s  %s\n%s  %s", csTar, extensionFileName, csjs, extension.ExtensionJS)
	}
}

// rbacHandler returns extension RBAC configuration.
func rbacHandler(ext extension.Extension) gin.HandlerFunc {
	return createHandler(contentTypeYAML, ext.ProxyRBAC)
}
