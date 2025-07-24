((window) => {
    const component2 = (context, extensionName) => {
        const app = context.application;
        const appNamespace = app?.metadata?.namespace || '';
        const appName = app?.metadata?.name || '';
        const project = app?.spec?.project || '';
        const resource = context.resource;
        const resourceNamespace = resource?.metadata?.namespace || '';
        const resourceName = resource?.metadata?.name || '';

        const handleClick = async () => {
            try {
                const response = await fetch(`/extensions/touch-${extensionName}/v1/touch/${extensionName}/${resourceNamespace}/${resourceName}`, {
                    method: 'PUT',
                    headers: {
                        'cache-control': 'no-cache',
                        'Argocd-Application-Name': `${appNamespace}:${appName}`,
                        'Argocd-Project-Name': project,
                    }
                });
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
            } catch (error) {
                console.error('Error:', error);
            }
        };

        return React.createElement(
            "div",
            {},
            React.createElement(
                "button",
                {
                    onClick: handleClick,
                    className: "argo-button argo-button--base"
                },
                `Touch ${resource.kind}`
            )
        );
    };
    const component = ( extensionName ) => {
        return React.createElement("div", {}, `Hello World ${extensionName}`);
    };
    {{- range $name, $_ := .Resources }}
    const component_{{$name}} = (context) => {
        return component2(context, "{{$name}}");
    };
    {{- end }}
    {{- range $name, $res := .Resources }}
    window.extensionsAPI.registerResourceExtension(
        component_{{$name}},
        "{{ $res.Group }}",
        "{{ $res.Kind }}",
        "{{ if $res.UIExtension }}{{ $res.UIExtension.TabTitle }}{{ else }}Touch{{ end }}"
        {{- if and $res.UIExtension $res.UIExtension.Icon }},
        { icon: "{{$res.UIExtension.Icon}}" }{{ end }}
    );
    {{- end }}
})(window);