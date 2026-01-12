package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/bakito/argocd-touch-extension/internal/app"
	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/bakito/argocd-touch-extension/internal/version"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     version.Name,
		Short:   "ArgoCD Touch Extension",
		RunE:    runRoot,
		Version: version.Print(),
	}
	configFile        string
	serviceAddress    string
	extensionTemplate string
	debug             bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.ErrorContext(context.Background(), "Error running application", "error", err)
		os.Exit(1)
	}
}

func init() {
	initConfigFlags(rootCmd)
}

func initConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&extensionTemplate, "extension-template", "", "Allows overwriting the UI extension template.")
	cmd.Flags().
		StringVar(&serviceAddress, "service-address", "http://argo-cd-touch-extension.svc.cluster.local:8080", "Service address")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Location of the config file")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
	_ = cmd.MarkFlagRequired("config")
}

func runRoot(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	application, err := app.New(cmd.Context(), cfg)
	if err != nil {
		return err
	}

	return application.Run(cmd.Context(), debug)
}

func loadConfig() (config.TouchConfig, error) {
	cfg, err := config.Load(configFile)
	if err != nil {
		return config.TouchConfig{}, err
	}
	cfg.ServiceAddress = serviceAddress
	cfg.ExtensionTemplate = extensionTemplate
	return cfg, nil
}
