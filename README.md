# ArgoCD Extension to touch Resources
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/argocd-touch-extension)](https://artifacthub.io/packages/search?repo=argocd-touch-extension)


An ArgoCD extension, enabling to touch resources by adding an annotation.
This is useful for e.g. ArgoCD to trigger a re-sync of a resource.

Resources can be configured individually, where each resource will receive an additional tab in the Resource detail
view.

This can be helpful to trigger
refresh [ExternalSecrets](https://external-secrets.io/latest/introduction/faq/#can-i-manually-trigger-a-secret-refresh).

## Config

Config for ArgoCD can be generated. Use `argocd-touch-extension config --help` for options.

## Links

- [UI Extensions](https://argo-cd.readthedocs.io/en/stable/developer-guide/extensions/ui-extensions/)
- [Proxy Extensions](https://argo-cd.readthedocs.io/en/stable/developer-guide/extensions/proxy-extensions/)
- [Argo CD Extension Installer](https://github.com/argoproj-labs/argocd-extension-installer)
