document.addEventListener('DOMContentLoaded', () => {
    // 1. Initialize Lucide Icons
    lucide.createIcons();

    // 2. Dark/Light Theme Toggle
    const themeToggle = document.getElementById('themeToggle');
    const htmlEl = document.documentElement;

    // Load saved theme
    const savedTheme = localStorage.getItem('theme') || 'dark';
    htmlEl.setAttribute('data-theme', savedTheme);

    if (themeToggle) {
        themeToggle.addEventListener('click', () => {
            const currentTheme = htmlEl.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            htmlEl.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
        });
    }

    // 3. Mobile Navigation Toggle
    const nav = document.querySelector('.nav');
    const toggle = document.querySelector('.nav-toggle');

    if (toggle && nav) {
        toggle.addEventListener('click', () => {
            const isOpen = nav.getAttribute('data-open') === 'true';
            nav.setAttribute('data-open', String(!isOpen));
            toggle.setAttribute('aria-expanded', String(!isOpen));
        });

        // Close on clicking outside
        document.addEventListener('click', (event) => {
            if (!nav.contains(event.target) && !toggle.contains(event.target)) {
                nav.setAttribute('data-open', 'false');
                toggle.setAttribute('aria-expanded', 'false');
            }
        });
    }

    // 4. Stat Counter Animation
    const stats = document.querySelectorAll('.stat-number');
    const animateStats = () => {
        stats.forEach(stat => {
            const target = +stat.getAttribute('data-target');
            let count = 0;
            const speed = target / 100;
            const updateCount = () => {
                count += speed;
                if (count < target) {
                    stat.innerText = Math.floor(count).toLocaleString();
                    setTimeout(updateCount, 15);
                } else {
                    stat.innerText = target.toLocaleString();
                }
            };
            updateCount();
        });
    };

    // Trigger stats animation when visible
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                animateStats();
                observer.unobserve(entry.target);
            }
        });
    }, { threshold: 0.5 });

    const statsSection = document.querySelector('.hero-stats');
    if (statsSection) {
        observer.observe(statsSection);
    }

    // 5. Searchable FAQ Accordion
    const faqSearchInput = document.getElementById('faqSearchInput');
    const faqItems = document.querySelectorAll('.faq-item');

    if (faqSearchInput) {
        faqSearchInput.addEventListener('input', (e) => {
            const query = e.target.value.toLowerCase().trim();
            faqItems.forEach(item => {
                const question = item.querySelector('.faq-question').textContent.toLowerCase();
                const answer = item.querySelector('.faq-answer').textContent.toLowerCase();
                if (question.includes(query) || answer.includes(query)) {
                    item.style.display = 'block';
                } else {
                    item.style.display = 'none';
                }
            });
        });
    }

    // FAQ Accordion Toggle
    faqItems.forEach(item => {
        const questionBtn = item.querySelector('.faq-question');
        questionBtn.addEventListener('click', () => {
            const isOpen = item.getAttribute('data-open') === 'true';
            
            // Close other items
            faqItems.forEach(i => {
                i.setAttribute('data-open', 'false');
                i.querySelector('.faq-question').setAttribute('aria-expanded', 'false');
            });

            if (!isOpen) {
                item.setAttribute('data-open', 'true');
                questionBtn.setAttribute('aria-expanded', 'true');
            }
        });
    });

    // 6. Interactive Terminal Simulator
    const terminalOutput = document.getElementById('terminalOutput');
    const terminalStatus = document.getElementById('terminalStatus');

    if (terminalOutput && terminalStatus) {
        const lines = [
            { text: '\nConnecting to gateway.proxvn.com:8882...', delay: 600, class: 'terminal-yellow' },
            { text: 'TLS Handshake successful.', delay: 400, class: 'terminal-green' },
            { text: 'Verifying credentials with server...', delay: 500, class: 'terminal-yellow' },
            { text: 'Authenticated as user "demo_user".\n', delay: 400, class: 'terminal-green' },
            { text: 'State updated: CONNECTED', delay: 300, class: 'terminal-purple' },
            { text: 'Generation ID: 1729 | Active Tunnel Session established.\n', delay: 200, class: 'terminal-blue' },
            { text: 'Tunnel URL: \x1b[32mhttps://shop-local.proxvn.dev\x1b[0m', delay: 400, class: 'terminal-green' },
            { text: 'Routing: localhost:3000 <-> https://shop-local.proxvn.dev\n', delay: 200, class: 'terminal-blue' },
            { text: 'Waiting for requests...', delay: 800, class: '' },
            { text: 'GET /api/v1/products - 200 OK (14ms)', delay: 1200, class: 'terminal-green' },
            { text: 'GET /static/css/main.css - 200 OK (4ms)', delay: 400, class: 'terminal-green' },
            { text: 'POST /api/v1/cart/add - 201 Created (48ms)', delay: 1800, class: 'terminal-green' },
            { text: 'GET /api/v1/checkout - 200 OK (22ms)', delay: 1000, class: 'terminal-green' },
            { text: 'Connection lost. Retrying to connect...', delay: 2000, class: 'terminal-yellow' },
            { text: 'Reconnected! Reclaiming Session (Generation ID: 1730)...', delay: 1000, class: 'terminal-green' },
            { text: 'Session successfully reclaimed. Subdomain kept: shop-local.proxvn.dev', delay: 500, class: 'terminal-green' },
        ];

        let lineIndex = 0;
        const runTerminalSimulation = () => {
            if (lineIndex < lines.length) {
                const line = lines[lineIndex];
                const span = document.createElement('div');
                if (line.class) span.className = line.class;
                
                // Replace ESC colors
                let cleanText = line.text.replace(/\x1b\[32m/g, '').replace(/\x1b\[0m/g, '');
                span.innerText = cleanText;

                // Adjust status text
                if (cleanText.includes('Connecting')) {
                    terminalStatus.innerText = 'CONNECTING';
                    terminalStatus.style.backgroundColor = 'rgba(245, 158, 11, 0.15)';
                    terminalStatus.style.color = '#f59e0b';
                } else if (cleanText.includes('State updated: CONNECTED')) {
                    terminalStatus.innerText = 'ACTIVE';
                    terminalStatus.style.backgroundColor = 'rgba(16, 185, 129, 0.15)';
                    terminalStatus.style.color = '#10b981';
                } else if (cleanText.includes('Connection lost')) {
                    terminalStatus.innerText = 'RECONNECTING';
                    terminalStatus.style.backgroundColor = 'rgba(239, 68, 68, 0.15)';
                    terminalStatus.style.color = '#ef4444';
                } else if (cleanText.includes('Reconnected!')) {
                    terminalStatus.innerText = 'ACTIVE';
                    terminalStatus.style.backgroundColor = 'rgba(16, 185, 129, 0.15)';
                    terminalStatus.style.color = '#10b981';
                }

                terminalOutput.appendChild(span);
                terminalOutput.scrollTop = terminalOutput.scrollHeight;
                
                lineIndex++;
                setTimeout(runTerminalSimulation, line.delay);
            } else {
                // Loop simulation
                setTimeout(() => {
                    terminalOutput.innerHTML = '';
                    lineIndex = 0;
                    runTerminalSimulation();
                }, 4000);
            }
        };

        setTimeout(runTerminalSimulation, 800);
    }
});
