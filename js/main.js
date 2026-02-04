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

    // Initialize search
    initSearch();
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

// Search functionality
const searchData = [
    { title: 'Overview', url: 'index.html', keywords: 'home overview introduction getting started' },
    { title: 'Installation', url: 'pages/getting-started.html', keywords: 'install setup go get environment variables configuration' },
    { title: 'Authentication', url: 'pages/authentication.html', keywords: 'oauth login token refresh password credentials' },
    { title: 'SObjects API', url: 'pages/sobjects.html', keywords: 'sobject crud create read update delete record' },
    { title: 'Query (SOQL)', url: 'pages/query.html', keywords: 'query soql select where builder' },
    { title: 'Search (SOSL)', url: 'pages/search.html', keywords: 'search sosl find builder' },
    { title: 'Bulk API 2.0', url: 'pages/bulk.html', keywords: 'bulk batch import export job csv' },
    { title: 'Composite API', url: 'pages/composite.html', keywords: 'composite batch tree graph subrequest' },
    { title: 'Analytics', url: 'pages/analytics.html', keywords: 'analytics reports dashboards' },
    { title: 'Tooling API', url: 'pages/tooling.html', keywords: 'tooling apex metadata deploy test' },
    { title: 'Chatter API', url: 'pages/chatter.html', keywords: 'chatter feed post comment connect social' },
    { title: 'UI API', url: 'pages/uiapi.html', keywords: 'ui layout picklist record form' },
    { title: 'Limits', url: 'pages/limits.html', keywords: 'limits api usage quota monitoring' },
    { title: 'Apex REST', url: 'pages/apex.html', keywords: 'apex rest custom endpoint webservice' },
    { title: 'Examples', url: 'pages/examples.html', keywords: 'example code sample tutorial guide' },
    { title: 'Error Handling', url: 'pages/errors.html', keywords: 'error handling exception retry debug' }
];

function initSearch() {
    const searchInput = document.getElementById('search-input');
    const searchResults = document.getElementById('search-results');

    if (!searchInput || !searchResults) return;

    searchInput.addEventListener('input', (e) => {
        const query = e.target.value.toLowerCase().trim();

        if (query.length < 2) {
            searchResults.style.display = 'none';
            return;
        }

        const results = searchData.filter(item =>
            item.title.toLowerCase().includes(query) ||
            item.keywords.toLowerCase().includes(query)
        );

        if (results.length > 0) {
            searchResults.innerHTML = results.map(item => {
                const basePath = window.location.pathname.includes('/pages/') ? '../' : '';
                return `<a href="${basePath}${item.url}" class="search-result-item">${item.title}</a>`;
            }).join('');
            searchResults.style.display = 'block';
        } else {
            searchResults.innerHTML = '<div class="search-no-results">No results found</div>';
            searchResults.style.display = 'block';
        }
    });

    // Close search on click outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.search-container')) {
            searchResults.style.display = 'none';
        }
    });

    // Keyboard navigation
    searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            searchResults.style.display = 'none';
            searchInput.blur();
        }
    });
}
