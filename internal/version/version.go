package version

import (
	"fmt"
)

const (
	Name            = "argocd-touch-extension"
	versionInfoTmpl = "%s (build: %s)"
)

// Build information. Populated at build-time.
var (
	Version = "dev"
	Build   = "N/A"
)

// Print returns version information.
func Print() string {
	return fmt.Sprintf(versionInfoTmpl, Version, Build)
}
