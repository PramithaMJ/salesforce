// Copy to clipboard functionality
function copyCode(btn) {
    const codeBlock = btn.closest('.code-block');
    const code = codeBlock.querySelector('code').textContent;

    navigator.clipboard.writeText(code).then(() => {
        const originalText = btn.textContent;
        btn.textContent = 'Copied!';
        btn.style.color = '#3fb950';
        setTimeout(() => {
            btn.textContent = originalText;
            btn.style.color = '';
        }, 2000);
    });
}

// Toggle API cards
document.querySelectorAll('.api-header').forEach(header => {
    header.addEventListener('click', () => {
        header.closest('.api-card').classList.toggle('open');
    });
});

// Active navigation
document.addEventListener('DOMContentLoaded', () => {
    const currentPath = window.location.pathname;
    document.querySelectorAll('.sidebar-links a').forEach(link => {
        if (link.getAttribute('href') === currentPath ||
            (currentPath.endsWith('/') && link.getAttribute('href') === 'index.html')) {
            link.classList.add('active');
        }
    });
});

// Mobile menu toggle
function toggleMenu() {
    document.querySelector('.sidebar').classList.toggle('open');
}

// Smooth scroll for anchor links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            target.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    });
});
