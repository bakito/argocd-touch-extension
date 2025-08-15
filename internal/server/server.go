package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	headerArgocdAppName       = "Argocd-Application-Name"
	headerArgocdProjName      = "Argocd-Project-Name"
	headerArgocdExtensionName = "Argocd-Touch-Extension-Name"
	headerArgoCDUsername      = "Argocd-Username"

	APIPathV1        = "/v1"
	apiPatchTouch    = "/touch"
	APIPathExtension = "/extension/"
)

func Run(ctx context.Context, client k8s.Client, ext extension.Extension, debug bool) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "argocd-touch-extension")
	})

	v1 := router.Group(APIPathV1)
	if debug {
		v1.Use(sloggin.New(slog.Default()))
	}

	v1Ext := v1.Group(APIPathExtension)

	v1Ext.GET(extension.ExtensionJS, jsHandler(ext))
	v1Ext.GET(extensionFileName, tarHandler(ext))
	v1Ext.GET(ExtensionChecksum, tarChecksumHandler(ext))
	v1Ext.GET("config", configHandler(ext))
	v1Ext.GET("rbac", rbacHandler(ext))

	v1Touch := v1.Group(apiPatchTouch)
	v1Touch.Use(validateArgocdHeaders())

	for name, res := range ext.Resources() {
		slog.With("resource", name, "group", res.Group, "version", res.Version, "kind", res.Kind).
			InfoContext(ctx, "Registering handler")
		v1Touch.PUT(name+"/:namespace/:name", handleTouch(client, res))
	}

	return start(ctx, router)
}

func validateArgocdHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok, _ := validateHeader(c, headerArgocdAppName); !ok {
			return
		}
		if ok, _ := validateHeader(c, headerArgocdProjName); !ok {
			return
		}
		ok, extName := validateHeader(c, headerArgocdExtensionName)
		if !ok {
			return
		}
		if !strings.HasPrefix(c.Request.URL.Path, fmt.Sprintf("%s/%s/%s/", APIPathV1, apiPatchTouch, extName)) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid extension name: " + extName,
			})
			c.Abort()
		}
		c.Next()
	}
}

func validateHeader(c *gin.Context, name string) (bool, string) {
	header := c.GetHeader(name)
	if header == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required header: " + name,
		})
		c.Abort()
		return false, ""
	}
	return true, header
}

func start(ctx context.Context, router *gin.Engine) error {
	slog.With("port", ":8080").InfoContext(ctx, "Starting server")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "Error starting server", "error", err)
			quit <- syscall.SIGTERM
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.InfoContext(ctx, "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.InfoContext(ctx, "Server exiting")
	return nil
}

func handleTouch(cl k8s.Client, res config.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		l := slog.With("resource", res.Name, "namespace", namespace, "name", name)

		value := metav1.Now().Format(time.RFC3339)
		if user := c.GetHeader(headerArgoCDUsername); user != "" {
			value += " by: " + user
			l = l.With("user", user)
		}

		if err := cl.PatchAnnotation(c, res, namespace, name, "argocd.bakito.ch/touch", value); err != nil {
			l.ErrorContext(c, "Failed to touch resource", "error", err)
			var se *kerr.StatusError
			if errors.As(err, &se) {
				c.JSON(int(se.Status().Code), err)
				return
			}
			c.JSON(http.StatusBadRequest, err)
			return
		}
		l.InfoContext(c, "Resource touched")

		c.Status(http.StatusOK)
	}
}
