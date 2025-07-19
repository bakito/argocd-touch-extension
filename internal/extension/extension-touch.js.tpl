((window) => {
    const component = () => {
        return React.createElement("div", {}, "Hello World");
    };
    {{- range $_, $res := .Resources }}
    window.extensionsAPI.registerResourceExtension(
        component,
        "{{ $res.Group }}",
        "{{ $res.Kind }}",
        "Nice extension"
    );
    {{- end }}
})(window);