import { router } from './router.js';

// Function to handle navigation for SPA links
const navigateTo = (url) => {
    history.pushState(null, null, url);
    router();
};

// Listen for browser back/forward button clicks
window.addEventListener('popstate', router);

// Main application entry point
document.addEventListener('DOMContentLoaded', () => {
    // Add a global click listener to handle all navigation
    document.body.addEventListener('click', (e) => {
        // Find the closest 'a' tag with a 'data-link' attribute
        const link = e.target.closest('[data-link]');
        if (link) {
            e.preventDefault();
            navigateTo(link.href);
        }
    });

    // Initial route handling
    router();
});
