/**
 * Securities Trading Platform - Main JavaScript File
 */

// Global application object
window.TradingPlatform = {
    config: {
        apiBaseUrl: '/api',
        websocketUrl: 'ws://localhost:8080/ws',
        refreshInterval: 30000, // 30 seconds
        chartColors: {
            primary: '#0d6efd',
            success: '#198754',
            danger: '#dc3545',
            warning: '#ffc107',
            info: '#0dcaf0'
        }
    },
    
    // WebSocket connection
    ws: null,
    
    // Chart instances
    charts: {},
    
    // Initialize the application
    init: function() {
        console.log('Initializing Trading Platform...');
        
        // Initialize components
        this.initTooltips();
        this.initModals();
        this.initFormValidation();
        this.initWebSocket();
        this.initAutoRefresh();
        this.initTheme();
        
        console.log('Trading Platform initialized successfully');
    },
    
    // Initialize Bootstrap tooltips
    initTooltips: function() {
        const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
        tooltipTriggerList.map(function(tooltipTriggerEl) {
            return new bootstrap.Tooltip(tooltipTriggerEl);
        });
    },
    
    // Initialize Bootstrap modals
    initModals: function() {
        // Auto-focus first input in modals
        document.addEventListener('shown.bs.modal', function(event) {
            const firstInput = event.target.querySelector('input, select, textarea');
            if (firstInput) {
                firstInput.focus();
            }
        });
    },
    
    // Initialize form validation
    initFormValidation: function() {
        // Add Bootstrap validation classes
        const forms = document.querySelectorAll('.needs-validation');
        Array.prototype.slice.call(forms).forEach(function(form) {
            form.addEventListener('submit', function(event) {
                if (!form.checkValidity()) {
                    event.preventDefault();
                    event.stopPropagation();
                }
                form.classList.add('was-validated');
            }, false);
        });
        
        // Real-time validation feedback
        const inputs = document.querySelectorAll('input[required], select[required], textarea[required]');
        inputs.forEach(function(input) {
            input.addEventListener('blur', function() {
                if (this.checkValidity()) {
                    this.classList.remove('is-invalid');
                    this.classList.add('is-valid');
                } else {
                    this.classList.remove('is-valid');
                    this.classList.add('is-invalid');
                }
            });
        });
    },
    
    // Initialize WebSocket connection for real-time updates
    initWebSocket: function() {
        if (!window.WebSocket) {
            console.warn('WebSocket not supported');
            return;
        }
        
        try {
            this.ws = new WebSocket(this.config.websocketUrl);
            
            this.ws.onopen = function(event) {
                console.log('WebSocket connected');
                TradingPlatform.showToast('Connected to real-time updates', 'success');
            };
            
            this.ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                TradingPlatform.handleWebSocketMessage(data);
            };
            
            this.ws.onclose = function(event) {
                console.log('WebSocket disconnected');
                // Attempt to reconnect after 5 seconds
                setTimeout(function() {
                    TradingPlatform.initWebSocket();
                }, 5000);
            };
            
            this.ws.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.warn('Failed to connect to WebSocket:', error);
        }
    },
    
    // Handle incoming WebSocket messages
    handleWebSocketMessage: function(data) {
        switch (data.type) {
            case 'trade_update':
                this.updateTradeStatus(data.payload);
                break;
            case 'market_data':
                this.updateMarketData(data.payload);
                break;
            case 'listing_update':
                this.updateListing(data.payload);
                break;
            case 'notification':
                this.showNotification(data.payload);
                break;
            default:
                console.log('Unknown WebSocket message type:', data.type);
        }
    },
    
    // Initialize auto-refresh for data
    initAutoRefresh: function() {
        // Refresh market data periodically
        if (document.querySelector('[data-auto-refresh]')) {
            setInterval(function() {
                TradingPlatform.refreshMarketData();
            }, this.config.refreshInterval);
        }
    },
    
    // Initialize theme handling
    initTheme: function() {
        // Detect system theme preference
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
        
        // Apply theme based on user preference or system setting
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme) {
            this.setTheme(savedTheme);
        } else if (prefersDark.matches) {
            this.setTheme('dark');
        }
        
        // Listen for system theme changes
        prefersDark.addEventListener('change', function(e) {
            if (!localStorage.getItem('theme')) {
                TradingPlatform.setTheme(e.matches ? 'dark' : 'light');
            }
        });
    },
    
    // Set application theme
    setTheme: function(theme) {
        document.documentElement.setAttribute('data-bs-theme', theme);
        localStorage.setItem('theme', theme);
        
        // Update theme toggle button if present
        const themeToggle = document.querySelector('[data-theme-toggle]');
        if (themeToggle) {
            const icon = themeToggle.querySelector('i');
            if (icon) {
                icon.className = theme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
            }
        }
    },
    
    // Toggle between light and dark themes
    toggleTheme: function() {
        const currentTheme = document.documentElement.getAttribute('data-bs-theme') || 'light';
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        this.setTheme(newTheme);
    },
    
    // Show toast notification
    showToast: function(message, type = 'info', options = {}) {
        const toastContainer = this.getToastContainer();
        
        const toastId = 'toast-' + Date.now();
        const toast = document.createElement('div');
        toast.id = toastId;
        toast.className = `toast align-items-center text-bg-${type} border-0`;
        toast.setAttribute('role', 'alert');
        toast.setAttribute('aria-live', 'assertive');
        toast.setAttribute('aria-atomic', 'true');
        
        const autoHide = options.autoHide !== false;
        const delay = options.delay || 5000;
        
        toast.innerHTML = `
            <div class="d-flex">
                <div class="toast-body">
                    ${options.icon ? `<i class="${options.icon} me-2"></i>` : ''}
                    ${message}
                </div>
                <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        `;
        
        toastContainer.appendChild(toast);
        
        const bsToast = new bootstrap.Toast(toast, {
            autohide: autoHide,
            delay: delay
        });
        
        bsToast.show();
        
        // Remove toast element after it's hidden
        toast.addEventListener('hidden.bs.toast', function() {
            toast.remove();
        });
        
        return toastId;
    },
    
    // Get or create toast container
    getToastContainer: function() {
        let container = document.getElementById('toast-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'toast-container';
            container.className = 'toast-container position-fixed bottom-0 end-0 p-3';
            container.style.zIndex = '1100';
            document.body.appendChild(container);
        }
        return container;
    },
    
    // Show notification modal
    showNotification: function(notification) {
        // Create notification badge if doesn't exist
        let badge = document.querySelector('.notification-badge');
        if (!badge) {
            badge = document.createElement('span');
            badge.className = 'notification-badge badge bg-danger position-absolute top-0 start-100 translate-middle rounded-pill';
            badge.style.fontSize = '0.7rem';
            
            const notificationIcon = document.querySelector('[data-notifications]');
            if (notificationIcon) {
                notificationIcon.appendChild(badge);
            }
        }
        
        // Update badge count
        const currentCount = parseInt(badge.textContent) || 0;
        badge.textContent = currentCount + 1;
        badge.style.display = 'block';
        
        // Show toast for important notifications
        if (notification.priority === 'high' || notification.priority === 'urgent') {
            this.showToast(notification.message, notification.priority === 'urgent' ? 'danger' : 'warning', {
                icon: 'fas fa-exclamation-triangle',
                autoHide: false
            });
        }
    },
    
    // API request helper
    apiRequest: function(endpoint, options = {}) {
        const url = this.config.apiBaseUrl + endpoint;
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'same-origin'
        };
        
        const requestOptions = { ...defaultOptions, ...options };
        
        if (requestOptions.body && typeof requestOptions.body === 'object') {
            requestOptions.body = JSON.stringify(requestOptions.body);
        }
        
        return fetch(url, requestOptions)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                return response.json();
            })
            .catch(error => {
                console.error('API request failed:', error);
                this.showToast(`Request failed: ${error.message}`, 'danger');
                throw error;
            });
    },
    
    // Refresh market data
    refreshMarketData: function() {
        const marketDataElements = document.querySelectorAll('[data-market-data]');
        
        marketDataElements.forEach(element => {
            const securityId = element.getAttribute('data-security-id');
            if (securityId) {
                this.apiRequest(`/market-data/${securityId}`)
                    .then(data => this.updateMarketDataElement(element, data))
                    .catch(error => console.error('Failed to refresh market data:', error));
            }
        });
    },
    
    // Update market data element
    updateMarketDataElement: function(element, data) {
        // Update price
        const priceElement = element.querySelector('[data-price]');
        if (priceElement && data.lastPrice) {
            const oldPrice = parseFloat(priceElement.textContent.replace(/[$,]/g, ''));
            const newPrice = data.lastPrice;
            
            priceElement.textContent = `$${newPrice.toFixed(2)}`;
            
            // Add animation for price changes
            if (oldPrice !== newPrice) {
                const changeClass = newPrice > oldPrice ? 'text-success' : 'text-danger';
                priceElement.classList.add(changeClass);
                setTimeout(() => priceElement.classList.remove(changeClass), 2000);
            }
        }
        
        // Update volume
        const volumeElement = element.querySelector('[data-volume]');
        if (volumeElement && data.volume) {
            volumeElement.textContent = data.volume.toLocaleString();
        }
        
        // Update change
        const changeElement = element.querySelector('[data-change]');
        if (changeElement && data.change !== undefined) {
            const changePercent = data.changePercent || 0;
            const changeText = `${data.change >= 0 ? '+' : ''}${data.change.toFixed(2)} (${changePercent.toFixed(2)}%)`;
            
            changeElement.textContent = changeText;
            changeElement.className = `badge ${data.change >= 0 ? 'bg-success' : 'bg-danger'}`;
        }
    },
    
    // Update trade status
    updateTradeStatus: function(trade) {
        const tradeRow = document.querySelector(`[data-trade-id="${trade.id}"]`);
        if (tradeRow) {
            const statusElement = tradeRow.querySelector('[data-status]');
            if (statusElement) {
                statusElement.innerHTML = this.getStatusBadge(trade.status);
            }
            
            const progressElement = tradeRow.querySelector('[data-progress]');
            if (progressElement && trade.progressPercent !== undefined) {
                progressElement.style.width = `${trade.progressPercent}%`;
                progressElement.setAttribute('aria-valuenow', trade.progressPercent);
            }
        }
    },
    
    // Get status badge HTML
    getStatusBadge: function(status) {
        const statusConfig = {
            'matched': { class: 'warning', text: 'Matched' },
            'confirmed': { class: 'info', text: 'Confirmed' },
            'settlement_initiated': { class: 'primary', text: 'Settlement' },
            'payment_received': { class: 'success', text: 'Payment' },
            'shares_transferred': { class: 'success', text: 'Transfer' },
            'settled': { class: 'success', text: 'Settled' },
            'failed': { class: 'danger', text: 'Failed' },
            'cancelled': { class: 'secondary', text: 'Cancelled' }
        };
        
        const config = statusConfig[status] || { class: 'secondary', text: status };
        return `<span class="badge bg-${config.class}">${config.text}</span>`;
    },
    
    // Format currency
    formatCurrency: function(amount, currency = 'USD') {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency
        }).format(amount);
    },
    
    // Format number with commas
    formatNumber: function(number) {
        return new Intl.NumberFormat('en-US').format(number);
    },
    
    // Format percentage
    formatPercentage: function(value, decimals = 2) {
        return `${value >= 0 ? '+' : ''}${value.toFixed(decimals)}%`;
    },
    
    // Debounce function for search inputs
    debounce: function(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = function() {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    window.TradingPlatform.init();
});

// Global utility functions
window.showToast = function(message, type, options) {
    return window.TradingPlatform.showToast(message, type, options);
};

window.toggleTheme = function() {
    window.TradingPlatform.toggleTheme();
};

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.TradingPlatform;
}