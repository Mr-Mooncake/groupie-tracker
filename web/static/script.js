// [Yasser] Live search: client -> server request on typing.
document.addEventListener('DOMContentLoaded', () => {
    const input = document.getElementById('search-input');
    const box = document.getElementById('search-results');
    if (!input || !box) return;

    const DEBOUNCE_MS = 250;
    const MIN_CHARS = 2;
    let timer = null;
    let controller = null;

    const clear = () => { box.replaceChildren(); };

    const showMessage = (text) => {
        const div = document.createElement('div');
        div.className = 'search-item';
        div.textContent = text;
        box.replaceChildren(div);
    };

    const render = (suggestions) => {
        if (!suggestions.length) return showMessage('No results found');
        const items = suggestions.map((s) => {
            const a = document.createElement('a');
            a.className = 'search-item';
            a.href = '/artist?id=' + encodeURIComponent(s.id);
            a.textContent = s.text; // textContent => no XSS
            return a;
        });
        box.replaceChildren(...items);
    };

    const search = async (query) => {
        if (controller) controller.abort(); // drop outdated request
        controller = new AbortController();
        try {
            const res = await fetch('/api/search?q=' + encodeURIComponent(query), {
                signal: controller.signal,
            });
            if (!res.ok) throw new Error('HTTP ' + res.status);
            render(await res.json());
        } catch (err) {
            if (err.name === 'AbortError') return;
            showMessage('Search unavailable, try again');
        }
    };

    input.addEventListener('input', () => {
        clearTimeout(timer);
        const query = input.value.trim();
        if (query.length < MIN_CHARS) {
            if (controller) controller.abort();
            clear();
            return;
        }
        timer = setTimeout(() => search(query), DEBOUNCE_MS);
    });

    document.addEventListener('click', (e) => {
        if (!e.target.closest('.search-box')) clear();
    });
    input.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') clear();
    });
});
