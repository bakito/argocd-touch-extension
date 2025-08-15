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
      image: quay.io/argoprojlabs/argocd-extension-installer:v0.0.8@sha256:e7cb054207620566286fce2d809b4f298a72474e0d8779ffa8ec92c3b630f054
      env:
        - name: EXTENSION_URL
          value: {{$.ServiceAddress}}/v1/extension/extension.tar.gz
        - name: EXTENSION_CHECKSUM_URL
          value: {{$.ServiceAddress}}/v1/extension/extension_checksum.txt
      volumeMounts:
        - name: extensions
          mountPath: /tmp/extensions/resources/
      securityContext:
        runAsUser: 999
        allowPrivilegeEscalation: false

  volumeMounts:
    - name: extensions
      mountPath: /tmp/extensions/touch/
  volumes:
    - name: extensions
      emptyDir: {}
