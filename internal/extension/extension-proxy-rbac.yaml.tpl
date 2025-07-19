apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: argocd-touch-proxy
  namespace: argo-cd
rules:
  {{- range $_, $res := .Resources }}
  - verbs:
      - get
      - patch
    apiGroups:
      - '{{$res.Group}}'
    resources:
      - '{{$res.Name}}'
  {{- end }}