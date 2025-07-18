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

func Run(client *dynamic.DynamicClient, configs map[string]TouchConfig) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	v1 := router.Group("/v1")
	v1.Use(validateArgocdHeaders())

	for resource, config := range configs {
		slog.With("resource", resource, "group", config.Group, "version", config.Version, "kind", config.Kind).
			Info("Registering handler")
		v1.PUT(resource+"/:namespace/:name", handleTouch(client, resource, config))
	}

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

func handleTouch(client *dynamic.DynamicClient, resource string, config TouchConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		cl := client.Resource(schema.GroupVersionResource{Group: config.Group, Version: config.Version, Resource: resource}).
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

type TouchConfig struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}
