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
                    let errorText = `Response was not ok (${response.status})`;
                    try {
                        const errorData = await response.text();
                        if (errorData) {
                            errorText = `${errorData} (${response.status})`;
                        }
                    } catch (e) {
                        // Fallback to default error message
                    }
                    const notifications = document.querySelector('.notifications-container');
                    if (notifications) {
                        const notification = document.createElement('div');
                        notification.className = 'argo-notification argo-notification--error';
                        notification.innerHTML = `<div class="argo-notification__content">${errorText}</div>`;
                        notifications.appendChild(notification);
                        setTimeout(() => notification.remove(), 5000);
                    }
                    throw new Error(errorText);
                }
            } catch (error) {
                console.error('Error:', error);
            }
        };

        return React.createElement(
            "div",
            {},
            React.createElement(
                "div",
                {},
                React.createElement(
                    "div",
                    {className: "argo-table-list"},
                    React.createElement(
                        "div",
                        {className: "argo-table-list__head"},
                        React.createElement("div", {className: "row"}, [
                            React.createElement("div", {className: "columns small-4"}, "Field"),
                            React.createElement("div", {className: "columns small-4"}, "Value"),
                            React.createElement("div", {className: "columns small-4"}, "Last")
                        ])
                    ),
                    React.createElement(
                        "div",
                        {className: "argo-table-list__row"},
                        React.createElement("div", {className: "row"}, [
                            React.createElement("div", {className: "columns small-4"}, "Last Touch"),
                            React.createElement("div", {className: "columns small-4"}, ""),
                            React.createElement("div", {className: "columns small-4"}, resource?.metadata?.annotations?.['argocd.bakito.ch/touch'] || 'Never')
                        ])
                    ),
                    resource?.status?.conditions?.map(condition =>
                        React.createElement(
                            "div",
                            {className: "argo-table-list__row", key: condition.type},
                            React.createElement("div", {className: "row"}, [
                                React.createElement("div", {className: "columns small-4"}, condition.type),
                                React.createElement("div", {className: "columns small-4"}, condition.status),
                                React.createElement("div", {className: "columns small-4"}, condition.lastTransitionTime || '')
                            ])
                        )
                    )
                ),
                React.createElement(
                    "button",
                    {
                        onClick: handleClick,
                        className: "argo-button argo-button--base"
                    },
                    `Touch ${resource.kind}`
                )
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