package k8s

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bakito/argocd-touch-extension/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Client interface {
	PatchAnnotation(ctx context.Context, res config.Resource, namespace, name, annotationKey, annotationValue string) error
	SetNameAndVersion(resources map[string]config.Resource) (map[string]config.Resource, error)
}

type client struct {
	dynamic   dynamic.Interface
	discovery discovery.DiscoveryInterface
}

func NewClient(ctx context.Context) (Client, error) {
	clientCfg := ctrl.GetConfigOrDie()

	dynamicClient, err := dynamic.NewForConfig(clientCfg)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create dynamic client", "error", err)
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientCfg)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create discovery client", "error", err)
		return nil, err
	}

	return &client{
		dynamic:   dynamicClient,
		discovery: discoveryClient,
	}, nil
}

func (cl *client) SetNameAndVersion(resMap map[string]config.Resource) (map[string]config.Resource, error) {
	needsUpdate := false
	for _, res := range resMap {
		needsUpdate = needsUpdate || res.Version == "" || res.Name == ""
	}
	if !needsUpdate {
		return resMap, nil
	}
	resources, err := cl.discovery.ServerPreferredNamespacedResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server preferred resources: %w", err)
	}

	for key, res := range resMap {
		version, name, err := cl.GetNameAndVersion(resources, res.Group, res.Kind)
		if err != nil {
			return nil, err
		}
		if res.Version == "" {
			res.Version = version
		}
		res.Name = name
		resMap[key] = res
	}
	return resMap, nil
}

func (cl *client) GetNameAndVersion(resources []*metav1.APIResourceList, group, kind string) (name, version string, err error) {
	for _, list := range resources {
		if list == nil {
			continue
		}

		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		if gv.Group == group {
			for _, r := range list.APIResources {
				if r.Kind == kind {
					return gv.Version, r.Name, nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("no preferred version found for group %s and kind %s", group, kind)
}

func (cl *client) PatchAnnotation(ctx context.Context, res config.Resource, namespace, name, annotation, value string) error {
	rc := cl.dynamic.Resource(schema.GroupVersionResource{Group: res.Group, Version: res.Version, Resource: res.Name}).
		Namespace(namespace)

	_, err := rc.Patch(ctx,
		name, types.MergePatchType,
		[]byte(fmt.Sprintf(`{"metadata":{"annotations":{%q:%q}}}`, annotation, value)),
		metav1.PatchOptions{},
	)
	return err
}
