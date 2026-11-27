document.querySelectorAll('[data-dataset-select]').forEach(picker => {
    const select = picker.querySelector('select');
    const search = picker.querySelector('input[type="search"]');
    const status = picker.querySelector('[role="status"]');
    const initialOptions = Array.from(select.options, option => option.cloneNode(true));
    let timer;
    let controller;

    search.addEventListener('keydown', event => {
        if (event.key === 'Enter' && !event.isComposing) event.preventDefault();
    });

    search.addEventListener('input', () => {
        clearTimeout(timer);
        controller?.abort();
        controller = new AbortController();
        const { signal } = controller;
        const url = new URL(picker.dataset.datasetSelect, document.baseURI);
        url.searchParams.set('q', search.value.trim());
        status.textContent = 'Searching datasets…';
        timer = setTimeout(async () => {
            try {
                const response = await fetch(url, { signal, headers: { Accept: 'application/json' } });
                if (!response.ok) throw new Error('Dataset search failed');
                const choices = await response.json();
                if (signal.aborted) return;

                const selected = select.selectedOptions[0];
                const empty = select.options[0];
                const options = choices.map(({ id, name }) => new Option(name, id));
                if (selected?.value && !choices.some(({ id }) => id === selected.value)) {
                    options.unshift(selected);
                }
                select.replaceChildren(empty, ...options);
                select.value = selected?.value ?? '';
                status.textContent = choices.length === 0 ? 'No matching datasets.' : '';
            } catch (error) {
                if (!signal.aborted) status.textContent = 'Datasets could not be loaded.';
            }
        }, 250);
    });

    window.addEventListener('pageshow', event => {
        if (!event.persisted) return;
        clearTimeout(timer);
        controller?.abort();
        search.value = '';
        status.textContent = '';
        select.replaceChildren(...initialOptions.map(option => option.cloneNode(true)));
    });
});
