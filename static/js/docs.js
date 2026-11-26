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
        mermaid.initialize({
            startOnLoad: false,
            theme: matchMedia('(prefers-color-scheme: dark)').matches ? 'redux-dark-color' : 'redux-color',
            layout: 'elk',
            fontFamily: getComputedStyle(document.querySelector('article')).fontFamily,
        });
        mermaid.run();
    }

    addCopyButtons();
    renderMermaid();
})();
