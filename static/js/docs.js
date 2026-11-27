(function () {
    function addCopyButtons() {
        document.querySelectorAll('article pre:not(.mermaid)').forEach(function (pre) {
            const btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'docs-copy-button';
            btn.textContent = 'Copy';
            btn.addEventListener('click', async function () {
                const code = pre.querySelector('code') || pre;
                try {
                    await navigator.clipboard.writeText(code.textContent);
                    btn.textContent = 'Copied';
                } catch (err) {
                    btn.textContent = 'Copy failed';
                }
                setTimeout(function () {
                    btn.textContent = 'Copy';
                }, 1500);
            });
            pre.appendChild(btn);
        });
    }

    function renderMermaid() {
        // The Mermaid script is only included on pages with diagrams.
        if (!window.mermaid) {
            return;
        }
        const root = document.documentElement;
        const nodes = Array.from(document.querySelectorAll('article pre.mermaid'));
        const sources = nodes.map(node => node.textContent);
        const staging = document.createElement('div');
        staging.className = 'docs-diagram-staging';
        staging.setAttribute('aria-hidden', 'true');
        const fontFamily = getComputedStyle(document.querySelector('article')).fontFamily;
        let renderedTheme;
        let renderID = 0;
        let pending = Promise.resolve();

        function syncTheme() {
            pending = pending.then(async () => {
                const theme = root.dataset.theme === 'dark' ? 'redux-dark-color' : 'redux-color';
                if (theme === renderedTheme) {
                    return;
                }

                mermaid.initialize({ startOnLoad: false, theme, layout: 'elk', fontFamily });
                const diagrams = [];
                document.body.append(staging);
                try {
                    for (const source of sources) {
                        diagrams.push(await mermaid.render(`docs-mermaid-${++renderID}`, source, staging));
                    }
                } finally {
                    staging.remove();
                }
                if ((root.dataset.theme === 'dark') !== (theme === 'redux-dark-color')) {
                    return;
                }
                nodes.forEach((node, index) => {
                    const { svg, bindFunctions } = diagrams[index];
                    node.innerHTML = svg;
                    bindFunctions?.(node);
                });
                renderedTheme = theme;
            }).catch(error => console.error('Could not render Mermaid diagrams', error));
        }

        new MutationObserver(syncTheme).observe(root, { attributeFilter: ['data-theme'] });
        syncTheme();
    }

    const navigationToggle = document.querySelector('.docs-navigation-toggle');
    if (navigationToggle) {
        navigationToggle.dataset.ready = '';
        navigationToggle.addEventListener('click', () => {
            const open = navigationToggle.getAttribute('aria-expanded') === 'true';
            navigationToggle.setAttribute('aria-expanded', String(!open));
        });
    }

    addCopyButtons();
    renderMermaid();
})();
