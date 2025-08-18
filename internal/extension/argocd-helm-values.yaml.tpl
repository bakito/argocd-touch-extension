# ArgoCD Helm values

configs:
  rbac:
    "policy.csv": |-
    {{- range $name, $_ := .Resources }}
      p, role:admin,    extensions, invoke, touch-{{$name}}, allow
      p, role:readonly, extensions, invoke, touch-{{$name}}, deny
    {{- end }}

  params:
    "server.enable.proxy.extension": true
  cm:
    {{- range $name, $_ := .Resources }}
    "extension.config.touch-{{$name}}": |-
      services:
        - url: {{$.ServiceAddress}}
          headers:
            - name: Argocd-Touch-Extension-Name
              value: {{$name}}

    {{- end }}

server:
  initContainers:
    - name: extension-touch
      image: ghcr.io/bakito/argocd-touch-extension:{{$.Version}}
      args:
        - install
        # - "--graceful" # enable graceful error handling
      env:
        - name: EXTENSION_BASE_URL
          value: {{$.ServiceAddress}}
      volumeMounts:
        - name: tmp
          mountPath: /tmp
