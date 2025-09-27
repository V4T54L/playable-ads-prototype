import Auth from './auth.js';
import Layout from './components/Layout.js';
import Dashboard from './views/Dashboard.js';
import Project from './views/Project.js';
import Jobs from './views/Jobs.js';
import Analytics from './views/Analytics.js';
import Login from './views/Login.js';
import Register from './views/Register.js';
import NotFound from './views/NotFound.js';

// Define the routes and their corresponding views
// `requiresAuth` flag protects routes from unauthenticated access
const routes = [
    { path: '/', view: Dashboard, requiresAuth: true },
    { path: '/dashboard', view: Dashboard, requiresAuth: true },
    { path: '/project', view: Project, requiresAuth: true },
    { path: '/jobs', view: Jobs, requiresAuth: true },
    { path: '/analytics', view: Analytics, requiresAuth: true },
    { path: '/login', view: Login, requiresAuth: false },
    { path: '/register', view: Register, requiresAuth: false },
];

export const router = async () => {
    const app = document.getElementById('app');
    if (!app) return;

    // Find the matching route for the current URL path
    const potentialMatches = routes.map(route => {
        return {
            route: route,
            isMatch: window.location.pathname === route.path
        };
    });

    let match = potentialMatches.find(potentialMatch => potentialMatch.isMatch);

    // If no match is found, route to the 404 page
    if (!match) {
        match = {
            route: { path: '/404', view: NotFound, requiresAuth: false },
            isMatch: true
        };
    }

    // --- Auth Guard ---
    const isAuthenticated = Auth.isAuthenticated();
    const isAuthPage = ['/login', '/register'].includes(match.route.path);

    if (match.route.requiresAuth && !isAuthenticated) {
        history.pushState(null, null, '/login');
        return router(); // Re-run the router for the new path
    }

    if (isAuthPage && isAuthenticated) {
        history.pushState(null, null, '/dashboard');
        return router(); // Re-run the router for the new path
    }

    // Create an instance of the view
    const view = new match.route.view();
    
    let viewContent = await view.render();

    // Wrap authenticated views with the main layout (Navbar, etc.)
    if (match.route.requiresAuth) {
        app.innerHTML = Layout(viewContent);
    } else {
        app.innerHTML = viewContent;
    }
    
    // Attach any event listeners after the view has been rendered
    if (view.after_render) {
        await view.after_render();
    }

    // Handle logout button if it exists in the current view
    const logoutButton = document.getElementById('logout-button');
    if (logoutButton) {
        logoutButton.addEventListener('click', () => {
            Auth.logout();
            history.pushState(null, null, '/login');
            router();
        });
    }
};