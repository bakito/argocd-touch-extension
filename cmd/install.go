package cmd

import (
	"github.com/bakito/argocd-touch-extension/internal/install"
	"github.com/spf13/cobra"
)

var (
	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install the UI extension to argocd server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install.Do(cmd.Context())
		},
	}
)

func init() {
	rootCmd.AddCommand(installCmd)
}
