package cmd

import (
	"log/slog"
	"os"

	"github.com/bakito/argocd-touch-extension/internal/app"
	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "argocd-touch-extension",
		Short: "ArgoCD Touch Extension",
		RunE:  runRoot,
	}
	configFile     string
	serviceAddress string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error running application", "error", err)
		os.Exit(1)
	}
}

func init() {
	initConfigFlags(rootCmd)
}

func initConfigFlags(cmd *cobra.Command) {
	cmd.Flags().
		StringVar(&serviceAddress, "service-address", "http://argocd-touch-extension.svc.cluster.local:8080", "Service address")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Location of the config file")
	_ = cmd.MarkFlagRequired("config")
}

func runRoot(_ *cobra.Command, _ []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	application, err := app.New(cfg)
	if err != nil {
		return err
	}

	return application.Run()
}

func loadConfig() (config.TouchConfig, error) {
	cfg, err := config.Load(configFile)
	if err != nil {
		return config.TouchConfig{}, err
	}
	cfg.ServiceAddress = serviceAddress
	return cfg, nil
}
