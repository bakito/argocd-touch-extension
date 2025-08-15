---
# Kubernetes ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: argocd-touch-proxy
  namespace: argo-cd
rules:
  {{- range $group, $resources := .ResourcesByGroup }}
  - apiGroups:
      - '{{ $group }}'
    resources:
    {{- range $_, $res := $resources }}
      - {{$res}}
    {{- end }}
    verbs:
      - get
      - patch
{{- end }}
---
# Helm Chart Values config
rbac:
  rules:
    {{- range $group, $resources := .ResourcesByGroup }}
    - apiGroups:
        - '{{ $group }}'
      resources:
      {{- range $_, $res := $resources }}
        - {{$res}}
      {{- end }}
    {{- end }}