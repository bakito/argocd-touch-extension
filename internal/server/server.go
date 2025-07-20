package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/bakito/argocd-touch-extension/internal/k8s"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	headerArgocdAppName  = "Argocd-Application-Name"
	headerArgocdProjName = "Argocd-Project-Name"
)

func Run(client k8s.Client, ext extension.Extension, debug bool) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	if debug {
		router.Use(sloggin.New(slog.Default()))
	}
	router.Use(gin.Recovery())

	v1 := router.Group("/v1")
	v1.GET("extension/"+extensionFileName, tarHandler(ext))
	v1.GET("extension/"+extensionChecksum, tarChecksumHandler(ext))
	v1.GET("extension/config", configHandler(ext))
	v1.GET("extension/deployment", deploymentHandler(ext))
	v1.GET("extension/rbac", rbacHandler(ext))

	v1Touch := router.Group("/v1/touch")
	v1Touch.Use(validateArgocdHeaders())

	for name, res := range ext.Resources() {
		slog.With("resource", name, "group", res.Group, "version", res.Version, "kind", res.Kind).
			Info("Registering handler")
		v1Touch.PUT(name+"/:namespace/:name", handleTouch(client, res))
	}

	return start(router)
}

func validateArgocdHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Info("headers", "H", fmt.Sprintf("%v", c.Request.Header))
		if validateHeader(c, headerArgocdAppName) {
			return
		}
		if validateHeader(c, headerArgocdProjName) {
			return
		}
		c.Next()
	}
}

func validateHeader(c *gin.Context, name string) bool {
	if argocdApp := c.GetHeader(name); argocdApp == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required header: " + name,
		})
		c.Abort()
		return true
	}
	return false
}

func start(router *gin.Engine) error {
	slog.With("port", ":8080").Info("Starting server")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error starting server", "error", err)
			quit <- syscall.SIGTERM
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Server exiting")
	return nil
}

func handleTouch(cl k8s.Client, res config.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		if err := cl.PatchAnnotation(c, res, namespace, name, "argocd.bakito.ch/touch", metav1.Now().Format(time.RFC3339)); err != nil {
			slog.Error("Failed to patch annotation", "error", err, "resource", res.Name, "namespace", namespace, "name", name)
			var se *kerr.StatusError
			if errors.As(err, &se) {
				c.JSON(int(se.Status().Code), err)
				return
			}
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusOK)
	}
}
