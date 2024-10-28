htmx.defineExtension('json-enc', {
    onEvent: function (name, evt) {
        if (name === "htmx:configRequest") {
            evt.detail.headers['Content-Type'] = "application/json";
        }
    },
    
    encodeParameters: function(xhr, parameters, elt) {
        xhr.overrideMimeType('application/json');
        const nestedParams = {};

        function parseValue(value) {
            const floatValue = parseFloat(value);
            if (!isNaN(floatValue) && floatValue.toString() === value) {
                return floatValue;
            }

            const intValue = parseInt(value, 10);
            if (!isNaN(intValue) && intValue.toString() === value) {
                return intValue;
            }

            return value;
        }

        for (const [name, value] of Object.entries(parameters)) {
            const nameParts = name.split('.');
            let current = nestedParams;

            nameParts.forEach((part, index) => {
                if (!current[part]) {
                    current[part] = (index === nameParts.length - 1) ? parseValue(value) : {};
                }
                current = current[part];
            });
        }

        return JSON.stringify(nestedParams);
    }
});
