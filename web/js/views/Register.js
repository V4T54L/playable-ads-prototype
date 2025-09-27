import { apiRequest } from '../api.js';
import Auth from '../auth.js';
import { router } from '../router.js';

export default class {
    constructor() {
        document.title = 'Register';
    }

    async render() {
        return `
            <div class="min-h-screen flex items-center justify-center">
                <div class="max-w-md w-full bg-white p-8 rounded-xl shadow-lg">
                    <h2 class="text-3xl font-bold text-center text-gray-800 mb-6">Create an Account</h2>
                    <form id="register-form">
                        <div class="mb-4">
                            <label for="username" class="block text-sm font-medium text-gray-700 mb-1">Username</label>
                            <input type="text" id="username" name="username" required minlength="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                        </div>
                        <div class="mb-6">
                            <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
                            <input type="password" id="password" name="password" required minlength="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                        </div>
                        <button type="submit" class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg transition-colors">Register</button>
                        <p id="error-message" class="text-red-500 text-center mt-4 h-4"></p>
                    </form>
                    <p class="text-center text-sm text-gray-600 mt-6">
                        Already have an account? <a href="/login" class="font-medium text-blue-600 hover:underline" data-link>Login here</a>
                    </p>
                </div>
            </div>
        `;
    }

    async after_render() {
        const form = document.getElementById('register-form');
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = form.username.value;
            const password = form.password.value;
            const errorElement = document.getElementById('error-message');
            const button = form.querySelector('button');
            
            button.disabled = true;
            button.textContent = 'Registering...';
            errorElement.textContent = '';

            try {
                // Register the user
                await apiRequest('/auth/register', 'POST', { username, password });
                
                // Automatically log them in
                const data = await apiRequest('/auth/login', 'POST', { username, password });
                Auth.saveTokens(data.access_token, data.refresh_token);

                // Navigate to dashboard and re-run router
                history.pushState(null, null, '/dashboard');
                router();
            } catch (error) {
                errorElement.textContent = error.message;
                button.disabled = false;
                button.textContent = 'Register';
            }
        });
    }
}
