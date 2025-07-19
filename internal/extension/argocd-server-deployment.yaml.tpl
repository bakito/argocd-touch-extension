apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-server
spec:
  template:
    spec:
      initContainers:
        - name: extension-touch
          image: quay.io/argoprojlabs/argocd-extension-installer:<CHANGE-ME>>
          env:
          - name: EXTENSION_URL
            value: {{.ServiceAddress}}/v1/extension/tar
          volumeMounts:
            - name: extensions
              mountPath: /tmp/extensions/
          securityContext:
            runAsUser: 1000
            allowPrivilegeEscalation: false
      containers:
        - name: argocd-server
          volumeMounts:
            - name: extensions
              mountPath: /tmp/extensions/