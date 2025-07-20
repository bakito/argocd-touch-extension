package cmd

import (
	"fmt"

	"github.com/bakito/argocd-touch-extension/internal/app"
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Generate configuration files",
		RunE:  runConfig,
	}

	// flags.
	outputType string
)

func init() {
	rootCmd.AddCommand(configCmd)
	// Add flags
	initConfigFlags(configCmd)
	configCmd.Flags().StringVarP(&outputType, "type", "t", "all", "Output type (all, config, deployment, rbac, extension)")
}

func runConfig(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	application, err := app.New(cfg)
	if err != nil {
		return err
	}

	ext, err := application.Extension()
	if err != nil {
		return err
	}

	switch outputType {
	case "config":
		cmd.Println(string(ext.ArgoCDConfig()))
	case "deployment":
		cmd.Println(string(ext.ArgoCDDeployment()))
	case "rbac":
		cmd.Println(string(ext.ProxyRBAC()))
	case "extension":
		cmd.Println(string(ext.ExtensionJS()))
	case "all":
		cmd.Println("=== ArgoCD Config ===")
		cmd.Println(string(ext.ArgoCDConfig()))
		cmd.Println("\n=== ArgoCD Deployment ===")
		cmd.Println(string(ext.ArgoCDDeployment()))
		cmd.Println("\n=== RBAC Configuration ===")
		cmd.Println(string(ext.ProxyRBAC()))
		cmd.Println("\n=== JS Extension ===")
		cmd.Println(string(ext.ExtensionJS()))
	default:
		return fmt.Errorf("invalid output type: %s", outputType)
	}

	return nil
}
