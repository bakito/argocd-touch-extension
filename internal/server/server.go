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
	"github.com/gin-gonic/gin"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

const (
	headerArgocdAppName  = "Argocd-Application-Name"
	headerArgocdProjName = "Argocd-Project-Name"
)

func Run(client *dynamic.DynamicClient, ext extension.Extension) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	v1 := router.Group("/v1")
	v1.GET("extension/tar", handleExtensionTar(ext))
	v1.GET("extension/config", handleExtensionConfig(ext))
	v1.GET("extension/deployment", handleExtensionDeployment(ext))

	v1Touch := router.Group("/v1/touch")
	v1Touch.Use(validateArgocdHeaders())

	for name, res := range ext.Resources() {
		slog.With("resource", name, "group", res.Group, "version", res.Version, "kind", res.Kind).
			Info("Registering handler")
		v1Touch.PUT(name+"/:namespace/:name", handleTouch(client, name, res))
	}

	return start(router)
}

func validateArgocdHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
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

func handleTouch(client *dynamic.DynamicClient, resource string, res config.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		cl := client.Resource(schema.GroupVersionResource{Group: res.Group, Version: res.Version, Resource: resource}).
			Namespace(namespace)

		_, err := cl.Patch(
			c,
			name,
			types.MergePatchType,
			[]byte(
				fmt.Sprintf(`{"metadata":{"annotations":{"argocd.bakito.ch/touch":%q}}}`, metav1.Now().Format(time.RFC3339)),
			),
			metav1.PatchOptions{},
		)
		if err != nil {
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
