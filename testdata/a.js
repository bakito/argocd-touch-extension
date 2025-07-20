((window) => {
    const component2 = (context, extensionName) => {
        const handleClick = async () => {
            try {
                const response = await fetch(`/extensions/touch-${extensionName}`, {
                    method: 'PUT',
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
    const component = ( extensionName ) => {
        return React.createElement("div", {}, `Hello World ${extensionName}`);
    };
    const component_configmaps = (context) => {
        return component2(context, "configmaps");
    };
    const component_pods = (context) => {
        return component2(context, "pods");
    };
    const component_sa = (context) => {
        return component2(context, "sa");
    };
    window.extensionsAPI.registerResourceExtension(
        component_configmaps,
        "",
        "ConfigMap",
        "Touch"
    );
    window.extensionsAPI.registerResourceExtension(
        component_pods,
        "",
        "Pod",
        "Touch Pod",
        { icon: "fa-box" }
    );
    window.extensionsAPI.registerResourceExtension(
        component_sa,
        "",
        "ServiceAccount",
        "Touch SA",
        { icon: "fa-box" }
    );
})(window);

