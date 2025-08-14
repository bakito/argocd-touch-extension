package config

import (
	"fmt"
	"regexp"
)

var keyPattern = regexp.MustCompile("^[A-Za-z0-9_]{3,}$")

type TouchConfig struct {
	ServiceAddress    string
	ExtensionTemplate string
	Resources         Resources
}

type Resources map[string]Resource

func (r Resources) validateKeys() error {
	for key := range r {
		if !keyPattern.MatchString(key) {
			return fmt.Errorf("resource key %q must match pattern %q", key, keyPattern.String())
		}
	}
	return nil
}

type Resource struct {
	Group       string       `json:"group"                 yaml:"group"`
	Version     string       `json:"version"               yaml:"version"`
	Kind        string       `json:"kind"                  yaml:"kind"`
	Name        string       `json:"name"                  yaml:"name"`
	UIExtension *UIExtension `json:"uiExtension,omitempty" yaml:"uiExtension,omitempty"`
}

type UIExtension struct {
	TabTitle string `json:"tabTitle,omitempty" yaml:"tabTitle,omitempty"`
	Icon     string `json:"icon,omitempty"     yaml:"icon,omitempty"`
}
