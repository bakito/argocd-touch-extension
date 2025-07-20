((window) => {
    // Resource configurations
    const RESOURCE_CONFIG = {
        configMap: {
            kind: 'ConfigMap',
            title: 'Touch',
            icon: ''
        },
        pod: {
            kind: 'Pod',
            title: 'Touch Pod',
            icon: 'fa-box'
        }
    };

    // React components
    const ConfigMapView = () => {
        return React.createElement("div", {}, "Hello World configmaps");
    };

    const PodView = () => {
        return React.createElement("div", {}, "Hello World pods");
    };

    // Helper function for registering extensions
    const registerExtension = (component, config) => {
        window.extensionsAPI.registerResourceExtension(
            component,
            "",
            config.kind,
            config.title,
            config.icon ? { icon: config.icon } : undefined
        );
    };

    // Register extensions
    registerExtension(ConfigMapView, RESOURCE_CONFIG.configMap);
    registerExtension(PodView, RESOURCE_CONFIG.pod);
})(window);