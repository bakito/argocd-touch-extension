package main

import (
	"log/slog"
	"os"

	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/bakito/argocd-touch-extension/internal/server"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	client, err := dynamic.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		slog.Error("Failed to get ConfigMap", "error", err)
		os.Exit(1)
	}

	cfg := config.TouchConfig{
		ServiceAddress: "http://argocd-touch-extension.svc.cluster.local:8080",
		Resources: map[string]config.Resource{
			"configmaps": {
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap",
			},
			"pod": {
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
		},
	}

	ext, err := extension.New(cfg)
	if err != nil {
		slog.Error("Error sunning server", "error", err)
	}

	if err := server.Run(client, ext); err != nil {
		slog.Error("Error sunning server", "error", err)
	}
}
