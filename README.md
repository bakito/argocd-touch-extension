# ArgoCD Extension to touch Resources

https://argo-cd.readthedocs.io/en/stable/developer-guide/extensions/ui-extensions/
https://argo-cd.readthedocs.io/en/stable/developer-guide/extensions/proxy-extensions/

https://github.com/argoproj-labs/argocd-extension-installer

https://external-secrets.io/latest/introduction/faq/#can-i-manually-trigger-a-secret-refresh

https://github.com/argoproj-labs/argocd-ephemeral-access

https://github.com/opsmx-cnoe/argocd-extensions/blob/main/resources/extension-topmenu.js

## Setup

```bash
kind create cluster --config testdata/kind-config.yaml 

helm upgrade --install argo-cd -n argo-cd --create-namespace oci://ghcr.io/argoproj/argo-helm/argo-cd --version 8.1.3


docker build -t localhost:5000/argocd-touch-extension
docker push localhost:5000/argocd-touch-extension
```