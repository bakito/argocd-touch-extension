apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  extension.config: |
    extensions:
    {{- range $name, $_ := .Resources }}
    - name: touch-{{ $name }}
      backend:
        connectionTimeout: 2s
        keepAlive: 15s
        idleConnectionTimeout: 60s
        maxIdleConnections: 30
        services:
        - url: {{$.ServiceAddress}}/v1/touch/{{ $name }}
    {{- end }}