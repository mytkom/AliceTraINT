(function () {
    function parseEquals(raw) {
        try {
            return JSON.parse(raw);
        } catch (err) {
            return raw;
        }
    }

    function matchesEquals(actual, expected) {
        if (Array.isArray(expected)) {
            return expected.map(String).includes(String(actual));
        }
        if (expected === null || expected === undefined) {
            return actual === '' || actual === null || actual === undefined;
        }
        return String(actual) === String(expected);
    }

    function getConfigValue(form, fieldPath) {
        const name = 'configuration.' + fieldPath;
        const boolCheckbox = form.querySelector('input[type="checkbox"][data-bool-target="' + name + '"]');
        if (boolCheckbox) {
            return boolCheckbox.checked;
        }
        const control = form.querySelector('[name="' + name + '"]');
        if (!control) {
            return undefined;
        }
        if (control.tagName === 'SELECT') {
            const selected = control.options[control.selectedIndex];
            if (selected && selected.getAttribute('data-null') === 'true') {
                return null;
            }
            return control.value;
        }
        return control.value;
    }

    function setGroupEnabled(group, enabled) {
        group.classList.toggle('hidden', !enabled);
        group.querySelectorAll('input, select, textarea').forEach(function (control) {
            const nestedOwner = control.closest('.nn-field-group');
            if (nestedOwner !== group && nestedOwner && nestedOwner.hasAttribute('data-show-when-field')) {
                return;
            }
            control.disabled = !enabled;
        });
    }

    function refreshShowWhen(form) {
        const groups = Array.from(form.querySelectorAll('.nn-field-group[data-show-when-field]'));
        groups.forEach(function (group) {
            const field = group.getAttribute('data-show-when-field');
            const expected = parseEquals(group.getAttribute('data-show-when-equals'));
            const actual = getConfigValue(form, field);

            let parentVisible = true;
            let parent = group.parentElement ? group.parentElement.closest('.nn-field-group') : null;
            while (parent) {
                if (parent.classList.contains('hidden')) {
                    parentVisible = false;
                    break;
                }
                parent = parent.parentElement ? parent.parentElement.closest('.nn-field-group') : null;
            }

            const visible = parentVisible && matchesEquals(actual, expected);
            setGroupEnabled(group, visible);
        });
    }

    function setControlValue(form, path, value) {
        const boolCheckbox = form.querySelector('input[type="checkbox"][data-bool-target="' + path + '"]');
        if (boolCheckbox) {
            boolCheckbox.checked = Boolean(value);
            if (window.AliceTraINTForm) {
                window.AliceTraINTForm.syncBoolHidden(boolCheckbox);
            }
            return true;
        }

        const rangeMin = form.querySelector('[data-range-parent="' + path + '"][data-range-index="0"]');
        const rangeMax = form.querySelector('[data-range-parent="' + path + '"][data-range-index="1"]');
        if (rangeMin && rangeMax && Array.isArray(value) && value.length >= 2) {
            rangeMin.value = value[0];
            rangeMax.value = value[1];
            return true;
        }

        const control = form.querySelector('[name="' + path + '"]');
        if (!control) {
            return false;
        }

        if (control.tagName === 'SELECT') {
            if (value === null || value === undefined) {
                const nullOpt = control.querySelector('option[data-null="true"]');
                if (nullOpt) {
                    nullOpt.selected = true;
                    return true;
                }
                control.value = '';
                return true;
            }
            control.value = String(value);
            return true;
        }

        const jsonType = control.getAttribute('data-json-type') || '';
        if ((jsonType === 'uint_list' || jsonType === 'string_list') && Array.isArray(value)) {
            control.value = value.join(', ');
            return true;
        }

        if (value === null || value === undefined) {
            control.value = '';
            return true;
        }
        control.value = String(value);
        return true;
    }

    function applyPrefillObject(form, obj, prefix) {
        if (!obj || typeof obj !== 'object' || Array.isArray(obj)) {
            return;
        }
        Object.keys(obj).forEach(function (key) {
            const path = prefix ? prefix + '.' + key : key;
            const value = obj[key];
            if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
                applyPrefillObject(form, value, path);
                return;
            }
            setControlValue(form, path, value);
        });
    }

    function applyPrefill(form) {
        const el = document.getElementById('prefill-task-config');
        if (!el) {
            return;
        }
        const raw = el.textContent.trim();
        if (!raw || raw === 'null') {
            return;
        }
        let config;
        try {
            config = JSON.parse(raw);
        } catch (err) {
            return;
        }
        if (!config || typeof config !== 'object') {
            return;
        }

        // Apply twice: selectors first (architecture/optimizer/...), then nested after groups shown.
        applyPrefillObject(form, config, 'configuration');
        refreshShowWhen(form);
        applyPrefillObject(form, config, 'configuration');
        refreshShowWhen(form);
        initBools(form);
    }

    function initBools(form) {
        form.querySelectorAll('input[type="checkbox"][data-bool-target]').forEach(function (checkbox) {
            if (window.AliceTraINTForm) {
                window.AliceTraINTForm.syncBoolHidden(checkbox);
            }
            checkbox.addEventListener('change', function () {
                if (window.AliceTraINTForm) {
                    window.AliceTraINTForm.syncBoolHidden(checkbox);
                }
            });
        });
    }

    function initForm(form) {
        initBools(form);
        refreshShowWhen(form);
        applyPrefill(form);
        form.addEventListener('change', function (event) {
            if (!event.target.matches('select, input')) {
                return;
            }
            if (event.target.matches('input[type="checkbox"][data-bool-target]') && window.AliceTraINTForm) {
                window.AliceTraINTForm.syncBoolHidden(event.target);
            }
            refreshShowWhen(form);
        });
        form.addEventListener('submit', function () {
            form.querySelectorAll('input[type="checkbox"][data-bool-target]').forEach(function (checkbox) {
                if (window.AliceTraINTForm) {
                    window.AliceTraINTForm.syncBoolHidden(checkbox);
                }
            });
        });
    }

    document.addEventListener('DOMContentLoaded', function () {
        const form = document.getElementById('create-training-task-form');
        if (form) {
            initForm(form);
        }
    });
})();
