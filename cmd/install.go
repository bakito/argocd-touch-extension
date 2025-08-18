package cmd

import (
	"log/slog"

	"github.com/bakito/argocd-touch-extension/internal/install"
	"github.com/spf13/cobra"
)

var (
	grace      bool
	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install the UI extension to argocd server",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := install.Do(cmd.Context())
			if err != nil && !grace {
				return err
			}
			slog.ErrorContext(cmd.Context(), "Extension installation failed, but continuing", "error", err)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVarP(&grace, "graceful", "g", false, "Continues normally if there is an error")
}
