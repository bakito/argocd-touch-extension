package main

import (
	"log/slog"
	"os"

	"github.com/bakito/argocd-touch-extension/internal/server"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	cfg := ctrl.GetConfigOrDie()

	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		slog.Error("Failed to get ConfigMap", "error", err)
		os.Exit(1)
	}

	configs := map[string]server.TouchConfig{
		"configmaps": {
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		},
	}

	if err := server.Run(client, configs); err != nil {
		slog.Error("Error sunning server", "error", err)
	}
}
