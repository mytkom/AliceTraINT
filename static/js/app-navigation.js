const menuButton = document.getElementById('menu-button');
const navigation = document.getElementById('app-navigation');
const themeButton = document.getElementById('theme-button');
const systemTheme = window.matchMedia('(prefers-color-scheme: dark)');

for (const link of navigation.querySelectorAll('a')) {
    if (window.location.pathname === link.pathname) {
        link.setAttribute('aria-current', 'page');
    } else if (window.location.pathname.startsWith(link.pathname + '/')) {
        link.setAttribute('aria-current', 'location');
    }
}

// JavaScript to toggle the mobile menu
menuButton.addEventListener('click', () => {
    const open = menuButton.getAttribute('aria-expanded') !== 'true';
    menuButton.setAttribute('aria-expanded', String(open));
    if (open) {
        navigation.querySelector('a').focus();
    }
});

navigation.closest('header').addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && menuButton.offsetWidth && menuButton.getAttribute('aria-expanded') === 'true') {
        menuButton.setAttribute('aria-expanded', 'false');
        menuButton.focus();
    }
});

function syncThemeButton() {
    const dark = document.documentElement.dataset.theme === 'dark';
    themeButton.setAttribute('aria-pressed', String(dark));
    themeButton.querySelector('.app-theme-icon').textContent = dark ? '☾' : '☀';
}

themeButton.addEventListener('click', () => {
    const theme = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
    document.documentElement.dataset.theme = theme;
    document.documentElement.dataset.themePreference = 'manual';
    try {
        localStorage.setItem('alicetraint-theme', theme);
    } catch {
        // The current page can still change theme when storage is unavailable.
    }
    syncThemeButton();
});

systemTheme.addEventListener('change', () => {
    if (document.documentElement.dataset.themePreference !== 'manual') {
        document.documentElement.dataset.theme = systemTheme.matches ? 'dark' : 'light';
        syncThemeButton();
    }
});

syncThemeButton();
