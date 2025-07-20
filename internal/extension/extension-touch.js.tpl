((window) => {
    const component = () => {
        return React.createElement("div", {}, "Hello World");
    };
    {{- range $_, $res := .Resources }}
    window.extensionsAPI.registerResourceExtension(
        component,
        "{{ $res.Group }}",
        "{{ $res.Kind }}",
        "{{ if $res.UIExtension }}{{ $res.UIExtension.TabTitle }}{{ else }}Touch{{ end }}"
        {{- if and $res.UIExtension $res.UIExtension.Icon }},
        { icon: "{{$res.UIExtension.Icon}}" }{{ end }}
    );
    {{- end }}
})(window);