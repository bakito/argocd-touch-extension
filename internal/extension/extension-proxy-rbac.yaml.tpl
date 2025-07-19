apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: argocd-touch-proxy
  namespace: argo-cd
rules:
  {{- range $group, $resources := .ResourcesByGroup }}
  - verbs:
      - get
      - patch
    apiGroups:
      - '{{ $group }}'
    resources:
    {{- range $_, $res := $resources }}
      - {{$res}}
    {{- end }}
  {{- end }}