// CSRF token injection for all HTMX requests
document.body.addEventListener('htmx:configRequest', function(event) {
    const csrfMeta = document.querySelector('meta[name="csrf-token"]');
    if (csrfMeta) {
        event.detail.headers['X-CSRF-Token'] = csrfMeta.content;
    }
});

// Handle HTMX response errors
document.body.addEventListener('htmx:responseError', function(event) {
    const status = event.detail.xhr.status;
    if (status === 401) {
        window.location.href = '/auth/login';
    } else if (status === 429) {
        showToast('Too many requests. Please slow down.', 'warning');
    }
});

// Handle HTMX redirect
document.body.addEventListener('htmx:beforeSwap', function(event) {
    if (event.detail.xhr.status === 422 || event.detail.xhr.status === 400) {
        event.detail.shouldSwap = true;
        event.detail.isError = false;
    }
});

// Toast notification helper
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    const colors = {
        info: 'bg-blue-500',
        success: 'bg-green-500',
        warning: 'bg-amber-500',
        error: 'bg-red-500'
    };

    toast.className = `${colors[type]} text-white px-4 py-3 rounded-lg shadow-lg toast-enter flex items-center gap-2`;
    toast.setAttribute('role', 'alert');
    toast.setAttribute('aria-live', 'polite');
    const span = document.createElement('span');
    span.textContent = message;
    toast.appendChild(span);

    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.remove('toast-enter');
        toast.classList.add('toast-exit');
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// Dark mode persistence (Alpine.js store)
document.addEventListener('alpine:init', () => {
    Alpine.store('app', {
        theme: localStorage.getItem('theme') || 'system',
        sidebarOpen: window.innerWidth >= 1024,
        unreadAlerts: 0,

        get isDark() {
            if (this.theme === 'system') {
                return window.matchMedia('(prefers-color-scheme: dark)').matches;
            }
            return this.theme === 'dark';
        },

        toggleTheme() {
            if (this.theme === 'dark') {
                this.theme = 'light';
            } else if (this.theme === 'light') {
                this.theme = 'system';
            } else {
                this.theme = 'dark';
            }
            localStorage.setItem('theme', this.theme);
            this.applyTheme();
        },

        applyTheme() {
            if (this.isDark) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
        },

        toggleSidebar() {
            this.sidebarOpen = !this.sidebarOpen;
        }
    });
});
