function copyCode() {
    const code = document.getElementById('quickstart-code');
    const text = code.textContent;

    navigator.clipboard.writeText(text).then(() => {
        const btn = document.querySelector('.copy-btn');
        btn.textContent = '✅ Copied!';
        setTimeout(() => {
            btn.textContent = '📋 Copy';
        }, 2000);
    });
}
