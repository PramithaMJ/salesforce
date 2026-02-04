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

// Comprehensive search data covering entire website
const searchData = [
    // Getting Started
    {
        title: 'Overview',
        url: 'index.html',
        category: 'Getting Started',
        description: 'Introduction to Salesforce Go SDK',
        keywords: 'home overview introduction sdk go golang salesforce api production-grade'
    },
    {
        title: 'Installation & Setup',
        url: 'pages/getting-started.html',
        category: 'Getting Started',
        description: 'How to install and configure the SDK',
        keywords: 'install setup go get environment variables configuration env SF_CLIENT_ID SF_CLIENT_SECRET SF_REFRESH_TOKEN SF_USERNAME SF_PASSWORD SF_ACCESS_TOKEN WithOAuthRefresh WithPasswordAuth WithAccessToken'
    },
    {
        title: 'Authentication',
        url: 'pages/authentication.html',
        category: 'Getting Started',
        description: 'OAuth 2.0, Password, and Token authentication',
        keywords: 'oauth oauth2 login token refresh password credentials authenticate connect security access token refresh token client id client secret username password security token'
    },

    // API Reference - Core
    {
        title: 'SObjects API',
        url: 'pages/sobjects.html',
        category: 'API Reference',
        description: 'Create, Read, Update, Delete SObject records',
        keywords: 'sobject sobjects crud create read update delete get patch post record account contact lead opportunity custom object describe metadata fields upsert external id'
    },
    {
        title: 'Query (SOQL)',
        url: 'pages/query.html',
        category: 'API Reference',
        description: 'Execute SOQL queries with builder pattern',
        keywords: 'query soql select from where and or like in limit offset order by asc desc builder pattern execute records result StringField IntField FloatField'
    },
    {
        title: 'Search (SOSL)',
        url: 'pages/search.html',
        category: 'API Reference',
        description: 'Full-text search with SOSL builder',
        keywords: 'search sosl find returning in all fields name email phone builder pattern'
    },

    // API Reference - Advanced
    {
        title: 'Bulk API 2.0',
        url: 'pages/bulk.html',
        category: 'API Reference',
        description: 'High-volume data operations with jobs',
        keywords: 'bulk batch import export job csv insert update upsert delete hard delete query ingest upload CreateJob CloseJob WaitForCompletion GetResults'
    },
    {
        title: 'Composite API',
        url: 'pages/composite.html',
        category: 'API Reference',
        description: 'Batch, Tree, Graph, Collections operations',
        keywords: 'composite batch tree graph subrequest sObject collections reference id allOrNone multi-operation transaction Execute'
    },
    {
        title: 'Analytics',
        url: 'pages/analytics.html',
        category: 'API Reference',
        description: 'Reports and Dashboards API',
        keywords: 'analytics reports dashboards run report list instances describe ReportResult DashboardResult'
    },
    {
        title: 'Tooling API',
        url: 'pages/tooling.html',
        category: 'API Reference',
        description: 'Apex, tests, metadata, debug logs',
        keywords: 'tooling apex metadata deploy test execute anonymous ApexClass ApexTrigger debug log code coverage'
    },
    {
        title: 'Chatter API',
        url: 'pages/chatter.html',
        category: 'API Reference',
        description: 'Social feeds, posts, comments, files',
        keywords: 'chatter connect feed post comment like mention user group file attachment social collaboration GetNewsFeed PostFeedElement'
    },
    {
        title: 'UI API',
        url: 'pages/uiapi.html',
        category: 'API Reference',
        description: 'Record layouts, picklists, form metadata',
        keywords: 'ui api layout picklist record form field metadata object info GetRecordUI GetPicklistValues GetLayout'
    },
    {
        title: 'Limits',
        url: 'pages/limits.html',
        category: 'API Reference',
        description: 'API usage and quota monitoring',
        keywords: 'limits api usage quota remaining monitoring DailyApiRequests PercentUsed Max GetLimits'
    },
    {
        title: 'Apex REST',
        url: 'pages/apex.html',
        category: 'API Reference',
        description: 'Call custom Apex REST endpoints',
        keywords: 'apex rest custom endpoint webservice @RestResource GetJSON PostJSON PatchJSON DeleteJSON'
    },

    // Resources
    {
        title: 'Examples',
        url: 'pages/examples.html',
        category: 'Resources',
        description: 'Real-world usage patterns and code samples',
        keywords: 'example code sample tutorial guide pattern real world common use case best practice'
    },
    {
        title: 'Error Handling',
        url: 'pages/errors.html',
        category: 'Resources',
        description: 'Handle API errors, retries, debugging',
        keywords: 'error handling exception retry debug troubleshoot IsNotFoundError IsAuthError IsRateLimitError rate limit 404 401 429 500'
    }
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
            item.keywords.toLowerCase().includes(query) ||
            item.description.toLowerCase().includes(query) ||
            item.category.toLowerCase().includes(query)
        );

        if (results.length > 0) {
            // Group by category
            const grouped = {};
            results.forEach(item => {
                if (!grouped[item.category]) {
                    grouped[item.category] = [];
                }
                grouped[item.category].push(item);
            });

            let html = '';
            const basePath = window.location.pathname.includes('/pages/') ? '../' : '';

            for (const [category, items] of Object.entries(grouped)) {
                html += `<div class="search-category">${category}</div>`;
                items.forEach(item => {
                    html += `
                        <a href="${basePath}${item.url}" class="search-result-item">
                            <div class="search-result-title">${highlightMatch(item.title, query)}</div>
                            <div class="search-result-desc">${item.description}</div>
                        </a>
                    `;
                });
            }

            searchResults.innerHTML = html;
            searchResults.style.display = 'block';
        } else {
            searchResults.innerHTML = '<div class="search-no-results">No results found for "' + escapeHtml(query) + '"</div>';
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
        if (e.key === 'Enter') {
            const firstResult = searchResults.querySelector('.search-result-item');
            if (firstResult) {
                window.location.href = firstResult.href;
            }
        }
    });

    // Global keyboard shortcut (Cmd/Ctrl + K)
    document.addEventListener('keydown', (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            searchInput.focus();
        }
    });
}

function highlightMatch(text, query) {
    const regex = new RegExp(`(${escapeRegex(query)})`, 'gi');
    return text.replace(regex, '<mark>$1</mark>');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function escapeRegex(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
