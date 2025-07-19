package k8s

import (
	"log/slog"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Clients struct {
	Dynamic   dynamic.Interface
	Discovery discovery.DiscoveryInterface
}

func NewClients() (*Clients, error) {
	clientCfg := ctrl.GetConfigOrDie()

	dynamicClient, err := dynamic.NewForConfig(clientCfg)
	if err != nil {
		slog.Error("Failed to create dynamic client", "error", err)
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientCfg)
	if err != nil {
		slog.Error("Failed to create discovery client", "error", err)
		return nil, err
	}

	return &Clients{
		Dynamic:   dynamicClient,
		Discovery: discoveryClient,
	}, nil
}
