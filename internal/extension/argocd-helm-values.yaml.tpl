# ArgoCD Helm values

configs:
  rbac:
    "policy.csv": |-
    {{- range $name, $_ := .Resources }}
      p, role:readonly, extensions, invoke, touch-{{$name}}, allow
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
      image: quay.io/argoprojlabs/argocd-extension-installer:v0.0.5@sha256:27e72f047298188e2de1a73a1901013c274c4760c92f82e6e46cd5fbd0957c6b
      env:
        - name: EXTENSION_URL
          value: {{$.ServiceAddress}}/v1/extension/extension.tar.gz
        - name: EXTENSION_CHECKSUM_URL
          value: {{$.ServiceAddress}}/v1/extension/extension_checksum.txt
      volumeMounts:
        - name: extensions
          mountPath: /tmp/extensions/resources/
      securityContext:
        runAsUser: 1000
        allowPrivilegeEscalation: false
  volumeMounts:
    - name: extensions
      mountPath: /tmp/extensions/touch/
  volumes:
    - name: extensions
      emptyDir: {}