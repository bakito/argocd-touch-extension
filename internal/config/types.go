package config

type TouchConfig struct {
	ServiceAddress string              `json:"serviceAddress"`
	Resources      map[string]Resource `json:"resources"`
}

type Resource struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}
