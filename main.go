package main

import (
	"log/slog"
	"os"

	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/bakito/argocd-touch-extension/internal/server"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	clientCfg := ctrl.GetConfigOrDie()
	client, err := dynamic.NewForConfig(clientCfg)
	if err != nil {
		slog.Error("Failed to create dynamic client", "error", err)
		os.Exit(1)
	}

	dcl, err := discovery.NewDiscoveryClientForConfig(clientCfg)
	if err != nil {
		slog.Error("Failed to create dynamic client", "error", err)
		os.Exit(1)
	}

	cfg := config.TouchConfig{
		ServiceAddress: "http://argocd-touch-extension.svc.cluster.local:8080",
		Resources: map[string]config.Resource{
			"configmaps": {
				Group: "",
				Kind:  "ConfigMap",
			},
			"pod": {
				Group: "",
				Kind:  "Pod",
			},
		},
	}

	ext, err := extension.New(cfg, dcl)
	if err != nil {
		slog.Error("Error sunning server", "error", err)
	}

	if err := server.Run(client, ext); err != nil {
		slog.Error("Error sunning server", "error", err)
	}
}
