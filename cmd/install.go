package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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

	httpTimeout = 30 * time.Second
)

func init() {
	rootCmd.AddCommand(installCmd)
}

func install(cmd *cobra.Command, _ []string) error {
	baseURL, ok := os.LookupEnv(EnvExtensionBaseURL)
	if !ok {
		return fmt.Errorf("missing environment variable: '%s'", EnvExtensionBaseURL)
	}
	extURL, err := url.JoinPath(baseURL, server.APIPathV1, server.APIPathExtension, extension.ExtensionJS)
	if err != nil {
		return fmt.Errorf("build extension URL: %w", err)
	}
	checksumURL, err := url.JoinPath(baseURL, server.APIPathV1, server.APIPathExtension, server.ExtensionChecksum)
	if err != nil {
		return fmt.Errorf("build checksum URL: %w", err)
	}
	// Download extension file
	extensionBytes, err := readAllFromURL(cmd.Context(), extURL)
	if err != nil {
		return fmt.Errorf("download extension: %w", err)
	}
	// Download checksum file
	checksumBytes, err := readAllFromURL(cmd.Context(), checksumURL)
	if err != nil {
		return fmt.Errorf("download checksum: %w", err)
	}
	// Calculate SHA256 of extension
	actualChecksum := checksumHex(extensionBytes)
	// Compare checksums
	expectedChecksum, found := extractChecksumFor(string(checksumBytes), extension.ExtensionJS)
	if !found {
		return fmt.Errorf("no checksum found for %s", extension.ExtensionJS)
	}
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch. Expected: %s, got: %s", expectedChecksum, actualChecksum)
	}
	cmd.Println(extURL, checksumURL)
	return nil
}

// readAllFromURL downloads the content at the given URL and returns the body as bytes.
func readAllFromURL(ctx context.Context, u string) ([]byte, error) {
	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, u)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// checksumHex returns the SHA256 hex string for the provided bytes.
func checksumHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// extractChecksumFor scans a checksum file content and returns the checksum for the target filename.
// It tolerates extra whitespace and ignores unrelated lines.
func extractChecksumFor(checksumFileContent, targetFile string) (string, bool) {
	lines := strings.Split(strings.TrimSpace(checksumFileContent), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		// Common formats: "<hex>  filename" or "<hex> filename"
		checkHex := fields[0]
		filename := fields[len(fields)-1]
		if filename == targetFile {
			return checkHex, true
		}
	}
	return "", false
}
