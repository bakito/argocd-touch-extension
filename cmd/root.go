package cmd

import (
	"log/slog"
	"os"

	"github.com/bakito/argocd-touch-extension/internal/app"
	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "argocd-touch-extension",
	Short: "ArgoCD Touch Extension",
	RunE:  runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error running application", "error", err)
		os.Exit(1)
	}
}

func runRoot(_ *cobra.Command, _ []string) error {
	cfg := config.TouchConfig{
		ServiceAddress: "http://argocd-touch-extension.svc.cluster.local:8080",
		Resources: map[string]config.Resource{
			"configmaps": {
				Group: "",
				Kind:  "ConfigMap",
			},
			"pods": {
				Group: "",
				Kind:  "Pod",
			},
		},
	}

	application, err := app.New(cfg)
	if err != nil {
		return err
	}

	return application.Run()
}
