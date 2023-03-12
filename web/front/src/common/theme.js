export function changeAppearance(appearance) {
    if (appearance === 'auto' || appearance === undefined) {
        appearance = window.matchMedia('(prefers-color-scheme: dark)').matches
            ? 'dark'
            : 'light';
    }
    document.body.setAttribute('data-theme', appearance);
    document
        .querySelector('meta[name="theme-color"]')
        .setAttribute('content', appearance === 'dark' ? '#222' : '#fff');
}