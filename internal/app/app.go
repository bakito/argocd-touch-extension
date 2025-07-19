package app

import (
	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/bakito/argocd-touch-extension/internal/k8s"
	"github.com/bakito/argocd-touch-extension/internal/server"
)

type Application struct {
	clients *k8s.Clients
	config  config.TouchConfig
}

func New(cfg config.TouchConfig) (*Application, error) {
	clients, err := k8s.NewClients()
	if err != nil {
		return nil, err
	}

	return &Application{
		clients: clients,
		config:  cfg,
	}, nil
}

func (a *Application) Run() error {
	ext, err := extension.New(a.config, a.clients.Discovery)
	if err != nil {
		return err
	}

	return server.Run(a.clients.Dynamic, ext)
}
