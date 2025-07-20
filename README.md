# ArgoCD Extension to touch Resources

https://argo-cd.readthedocs.io/en/stable/developer-guide/extensions/ui-extensions/
https://argo-cd.readthedocs.io/en/stable/developer-guide/extensions/proxy-extensions/

https://github.com/argoproj-labs/argocd-extension-installer

https://external-secrets.io/latest/introduction/faq/#can-i-manually-trigger-a-secret-refresh

https://github.com/argoproj-labs/argocd-ephemeral-access

https://github.com/opsmx-cnoe/argocd-extensions/blob/main/resources/extension-topmenu.js

## Setup

```bash

# install registry

docker start kind-registry || docker run -d --restart=always -p "127.0.0.1:5001:5000" --name kind-registry registry:3

docker build -t localhost:5001/argocd-touch-extension .
docker push localhost:5001/argocd-touch-extension

kind create cluster --config testdata/kind-config.yaml 

docker network connect kind kind-registry || true
kubectl create ns argo-cd

kubectl apply -f testdata/touch-config.yaml -f testdata/touch-deployment.yaml

helm upgrade --install argo-cd -n argo-cd oci://ghcr.io/argoproj/argo-helm/argo-cd -f testdata/argo-cd-values.yaml 

kubectl apply -f testdata/app.yaml

```