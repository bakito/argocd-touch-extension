package app

import (
	"context"

	"github.com/bakito/argocd-touch-extension/internal/config"
	"github.com/bakito/argocd-touch-extension/internal/extension"
	"github.com/bakito/argocd-touch-extension/internal/k8s"
	"github.com/bakito/argocd-touch-extension/internal/server"
)

type Application struct {
	client k8s.Client
	config config.TouchConfig
}

func New(ctx context.Context, cfg config.TouchConfig) (*Application, error) {
	client, err := k8s.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Application{
		client: client,
		config: cfg,
	}, nil
}

func (a *Application) Run(ctx context.Context, debug bool) error {
	ext, err := a.Extension()
	if err != nil {
		return err
	}

	return server.Run(ctx, a.client, ext, debug)
}

func (a *Application) Extension() (extension.Extension, error) {
	return extension.New(a.config, a.client, a.config.ExtensionTemplate)
}
