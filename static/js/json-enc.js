htmx.defineExtension('json-enc', {
    onEvent: function (name, evt) {
        if (name === "htmx:configRequest") {
            evt.detail.headers['Content-Type'] = "application/json";
        }
    },

    encodeParameters: function (xhr, parameters, elt) {
        xhr.overrideMimeType('application/json');
        const form = (elt && elt.closest) ? (elt.closest('form') || (elt.tagName === 'FORM' ? elt : null)) : null;
        if (!form || !window.AliceTraINTForm) {
            return JSON.stringify({});
        }
        return JSON.stringify(window.AliceTraINTForm.encodeForm(form));
    }
});

window.AliceTraINTForm = (function () {
    function setNested(root, path, value) {
        const parts = path.split('.');
        let current = root;
        for (let i = 0; i < parts.length; i++) {
            const part = parts[i];
            if (i === parts.length - 1) {
                current[part] = value;
                return;
            }
            if (!current[part] || typeof current[part] !== 'object' || Array.isArray(current[part])) {
                current[part] = {};
            }
            current = current[part];
        }
    }

    function parseList(value, jsonType) {
        const parts = String(value)
            .split(',')
            .map(function (part) { return part.trim(); })
            .filter(function (part) { return part.length > 0; });
        if (jsonType === 'uint_list') {
            return parts.map(function (part) { return parseInt(part, 10); });
        }
        return parts;
    }

    function parseScalar(value, jsonType) {
        if (jsonType === 'bool') {
            return value === true || value === 'true';
        }
        if (jsonType === 'select') {
            if (value === '' || value === undefined || value === null) {
                return null;
            }
            return value;
        }
        if (jsonType === 'uint' || jsonType === 'int') {
            const asInt = parseInt(value, 10);
            return isNaN(asInt) ? value : asInt;
        }
        if (jsonType === 'float64') {
            const asFloat = parseFloat(value);
            return isNaN(asFloat) ? value : asFloat;
        }
        if (jsonType === 'uint_list' || jsonType === 'string_list') {
            return parseList(value, jsonType);
        }
        return value;
    }

    function syncBoolHidden(checkbox) {
        const targetName = checkbox.getAttribute('data-bool-target');
        if (!targetName) {
            return;
        }
        const form = checkbox.closest('form');
        if (!form) {
            return;
        }
        const hidden = form.querySelector('input[type="hidden"][name="' + targetName + '"]');
        if (hidden) {
            hidden.value = checkbox.checked ? 'true' : 'false';
        }
    }

    function encodeForm(form) {
        form.querySelectorAll('input[type="checkbox"][data-bool-target]').forEach(syncBoolHidden);

        const nestedParams = {};
        const rangeParts = {};
        const controls = form.querySelectorAll('input[name], select[name], textarea[name]');

        controls.forEach(function (control) {
            if (control.disabled) {
                return;
            }

            if (control.type === 'checkbox') {
                return;
            }

            const jsonType = control.getAttribute('data-json-type') || '';
            const name = control.name;

            if (jsonType === 'float_range' || jsonType === 'int_range') {
                const parent = control.getAttribute('data-range-parent');
                const index = parseInt(control.getAttribute('data-range-index'), 10);
                if (!parent || isNaN(index)) {
                    return;
                }
                if (!rangeParts[parent]) {
                    rangeParts[parent] = { type: jsonType, values: [] };
                }
                const raw = control.value;
                rangeParts[parent].values[index] = jsonType === 'int_range'
                    ? parseInt(raw, 10)
                    : parseFloat(raw);
                return;
            }

            if (control.tagName === 'SELECT') {
                const selected = control.options[control.selectedIndex];
                if (jsonType === 'select') {
                    if (selected && selected.getAttribute('data-null') === 'true') {
                        setNested(nestedParams, name, null);
                        return;
                    }
                    setNested(nestedParams, name, parseScalar(control.value, 'select'));
                    return;
                }

                const raw = control.value;
                if (raw === '') {
                    return;
                }
                if (/^\d+$/.test(raw)) {
                    setNested(nestedParams, name, parseInt(raw, 10));
                    return;
                }
                setNested(nestedParams, name, raw);
                return;
            }

            setNested(nestedParams, name, parseScalar(control.value, jsonType));
        });

        Object.keys(rangeParts).forEach(function (name) {
            setNested(nestedParams, name, rangeParts[name].values);
        });

        return nestedParams;
    }

    return {
        encodeForm: encodeForm,
        syncBoolHidden: syncBoolHidden
    };
})();
