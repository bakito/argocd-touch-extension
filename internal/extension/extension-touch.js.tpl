((window) => {
    const component = ({ extensionName }) => {
        const handleClick = async () => {
            try {
                const response = await fetch(`/extensions/touch-${extensionName}`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
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
                { onClick: handleClick },
                "Touch Resource"
            )
        );
    };
    {{- range $name, $_ := .Resources }}
    const component_{{$name}} = () => {
        return React.createElement("div", {}, "Hello World {{$name}}");
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