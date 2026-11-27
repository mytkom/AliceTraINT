document.querySelectorAll('.task-list-date-field').forEach(field => {
    const text = field.querySelector('input[type="text"]');
    const calendar = field.querySelector('input[type="date"]');
    const button = field.querySelector('button');

    const syncCalendar = () => {
        calendar.value = text.validity.valid && text.value
            ? text.value.split('/').reverse().join('-')
            : '';
    };
    text.addEventListener('input', syncCalendar);
    calendar.addEventListener('change', () => {
        text.value = calendar.value.split('-').reverse().join('/');
    });
    window.addEventListener('pageshow', syncCalendar);
    syncCalendar();

    calendar.hidden = false;
    if (typeof calendar.showPicker !== 'function') return;

    field.classList.add('has-calendar-button');
    calendar.tabIndex = -1;
    calendar.setAttribute('aria-hidden', 'true');
    button.hidden = false;
    button.addEventListener('click', () => {
        syncCalendar();
        try {
            calendar.showPicker();
        } catch {
            field.classList.remove('has-calendar-button');
            calendar.removeAttribute('tabindex');
            calendar.removeAttribute('aria-hidden');
            button.hidden = true;
            calendar.focus();
        }
    });
});
