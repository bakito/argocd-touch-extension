package config

type TouchConfig struct {
	ServiceAddress    string
	ExtensionTemplate string
	Resources         map[string]Resource
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
