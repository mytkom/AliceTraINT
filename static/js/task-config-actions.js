(function () {
    function getRawConfig() {
        const el = document.getElementById('raw-task-config');
        if (!el) {
            return null;
        }
        const text = el.textContent.trim();
        if (!text || text === 'null') {
            return null;
        }
        return text;
    }

    document.addEventListener('DOMContentLoaded', function () {
        const code = document.getElementById('task-config-code');
        if (code && window.hljs) {
            window.hljs.highlightElement(code);
        }

        const copyBtn = document.getElementById('copy-task-config');
        if (!copyBtn) {
            return;
        }
        copyBtn.addEventListener('click', async function () {
            const raw = getRawConfig();
            if (!raw) {
                return;
            }
            try {
                let pretty = raw;
                try {
                    pretty = JSON.stringify(JSON.parse(raw), null, 2);
                } catch (err) {
                    // keep raw
                }
                await navigator.clipboard.writeText(pretty);
                const prev = copyBtn.textContent;
                copyBtn.textContent = 'Copied';
                setTimeout(function () {
                    copyBtn.textContent = prev;
                }, 1500);
            } catch (err) {
                copyBtn.textContent = 'Copy failed';
            }
        });
    });
})();
