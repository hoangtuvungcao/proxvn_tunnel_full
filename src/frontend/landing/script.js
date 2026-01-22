const nav = document.querySelector('.nav');
const toggle = document.querySelector('.nav-toggle');
const copyButtons = document.querySelectorAll('[data-target]');
const header = document.querySelector('.site-header');

if (toggle && nav) {
    toggle.addEventListener('click', () => {
        const isOpen = nav.getAttribute('data-open') === 'true';
        nav.setAttribute('data-open', String(!isOpen));
        toggle.setAttribute('aria-expanded', String(!isOpen));
    });

    document.addEventListener('click', (event) => {
        if (!nav.contains(event.target) && !toggle.contains(event.target)) {
            nav.setAttribute('data-open', 'false');
            toggle.setAttribute('aria-expanded', 'false');
        }
    });
}

if (copyButtons.length) {
    copyButtons.forEach((button) => {
        button.addEventListener('click', () => {
            const targetId = button.getAttribute('data-target');
            const code = document.getElementById(targetId);
            if (!code) return;

            navigator.clipboard.writeText(code.textContent.trim()).then(() => {
                const originalText = button.textContent;
                button.textContent = '✅ Copied!';
                setTimeout(() => {
                    button.textContent = originalText;
                }, 1800);
            });
        });
    });
}

if (header) {
    const observer = new IntersectionObserver(
        ([entry]) => {
            if (!entry.isIntersecting) {
                header.classList.add('is-solid');
            } else {
                header.classList.remove('is-solid');
            }
        },
        { rootMargin: '-120px 0px 0px 0px' }
    );

    observer.observe(document.body);
}
