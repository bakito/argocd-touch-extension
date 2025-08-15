package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"net/url"

	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/bakito/argocd-touch-extension/internal/server"
	"github.com/spf13/cobra"
)

const (
	EnvExtensionBaseURL         = "EXTENSION_BASE_URL"
	EnvExtensionInstallationDir = "EXTENSION_INSTALLATION_DIR"
)

var (
	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install the UI extension to argocd server",
		RunE:  install,
	}
)

func init() {
	rootCmd.AddCommand(installCmd)
}

func install(cmd *cobra.Command, _ []string) error {

	base, ok := os.LookupEnv(EnvExtensionBaseURL)
	if !ok {
		return fmt.Errorf("missing environment variable: '%s'", EnvExtensionBaseURL)
	}

	extensionURL, err := url.JoinPath(base, server.ApiPathV1, server.ApiPathExtension, extension.ExtensionJS)
	if err != nil {
		return err
	}
	checkSumURL, err := url.JoinPath(base, server.ApiPathV1, server.ApiPathExtension, server.ExtensionChecksum)
	if err != nil {
		return err
	}

	// Download extension file
	resp, err := http.Get(extensionURL)
	if err != nil {
		return fmt.Errorf("failed to download extension: %v", err)
	}
	defer resp.Body.Close()

	extensionBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read extension body: %v", err)
	}

	// Download checksum file
	respChecksum, err := http.Get(checkSumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksum: %v", err)
	}
	defer respChecksum.Body.Close()

	checksumBytes, err := io.ReadAll(respChecksum.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksum body: %v", err)
	}

	// Calculate SHA256 of extension
	hash := sha256.Sum256(extensionBytes)
	actualChecksum := hex.EncodeToString(hash[:])

	// Compare checksums
	expectedChecksum := strings.TrimSpace(string(checksumBytes))
	lines := strings.Split(expectedChecksum, "\n")
	foundChecksum := false
	for _, line := range lines {
		if strings.Contains(line, "extension-touch.js") {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				expectedChecksum = fields[0]
				if actualChecksum != expectedChecksum {
					return fmt.Errorf("checksum mismatch. Expected: %s, got: %s", expectedChecksum, actualChecksum)
				}
				foundChecksum = true
				break
			}
		}
	}
	if !foundChecksum {
		return fmt.Errorf("no checksum found for extension-touch.js")
	}
	if err != nil {
		return err
	}
	cmd.Println(extensionURL, checkSumURL)

	return nil
}
