document.querySelectorAll('[data-searchable-select]').forEach(picker => {
    const summary = picker.querySelector('summary');
    const label = summary.querySelector('span');
    const value = picker.querySelector('input[type="hidden"]');
    const search = picker.querySelector('input[type="search"]');
    const options = picker.querySelector('[data-select-options]');
    const status = picker.querySelector('[role="status"]');
    const initialValue = value.value;
    const initialLabel = label.textContent;
    const initialOptions = Array.from(options.children, option => option.cloneNode(true));
    let timer;
    let controller;

    function markSelection() {
        picker.querySelectorAll('button').forEach(button => {
            button.setAttribute('aria-pressed', button.value === value.value);
        });
    }

    picker.addEventListener('toggle', () => {
        if (picker.open) search.focus();
    });
    picker.addEventListener('click', event => {
        const button = event.target.closest('button');
        if (!button) return;
        value.value = button.value;
        label.textContent = button.textContent;
        markSelection();
        picker.open = false;
        summary.focus();
        value.dispatchEvent(new Event('change', { bubbles: true }));
    });
    picker.addEventListener('keydown', event => {
        if (event.key === 'Escape' && picker.open) {
            picker.open = false;
            summary.focus();
        }
        if (event.target === search && event.key === 'Enter' && !event.isComposing) {
            event.preventDefault();
        }
    });
    document.addEventListener('click', event => {
        if (!picker.contains(event.target)) picker.open = false;
    });

    search.addEventListener('input', () => {
        clearTimeout(timer);
        controller?.abort();
        controller = new AbortController();
        const { signal } = controller;
        const url = new URL(picker.dataset.searchableSelect, document.baseURI);
        url.searchParams.set('q', search.value.trim());
        status.textContent = 'Searching…';
        timer = setTimeout(async () => {
            try {
                const response = await fetch(url, { signal, headers: { Accept: 'application/json' } });
                if (!response.ok) throw new Error('Search failed');
                const choices = await response.json();
                if (signal.aborted) return;

                options.replaceChildren(...choices.map(({ id, name }) => {
                    const button = document.createElement('button');
                    button.type = 'button';
                    button.value = id;
                    button.textContent = name;
                    return button;
                }));
                markSelection();
                status.textContent = choices.length === 0 ? 'No matches.' : '';
            } catch {
                if (!signal.aborted) status.textContent = 'Choices could not be loaded.';
            }
        }, 250);
    });

    window.addEventListener('pageshow', event => {
        if (!event.persisted) return;
        clearTimeout(timer);
        controller?.abort();
        picker.open = false;
        value.value = initialValue;
        label.textContent = initialLabel;
        search.value = '';
        status.textContent = '';
        options.replaceChildren(...initialOptions.map(option => option.cloneNode(true)));
        markSelection();
    });
});
