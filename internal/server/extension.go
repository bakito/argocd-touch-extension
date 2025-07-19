package server

import (
	"net/http"

	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/gin-gonic/gin"
)

func handleExtensionConfig(ext extension.Extension) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", ext.ArgoCDConfig())
	}
}

func handleExtensionDeployment(ext extension.Extension) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", ext.ArgoCDDeployment())
	}
}

func handleExtensionTar(ext extension.Extension) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/x-tar", ext.ExtensionTar())
	}
}
