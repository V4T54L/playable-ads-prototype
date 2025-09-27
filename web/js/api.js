import Auth from './auth.js';
import { router } from './router.js';

const API_BASE_URL = ''; // Assuming API is served from the same origin

export const apiRequest = async (endpoint, method = 'GET', body = null, isFormData = false) => {
    const headers = {};
    const accessToken = Auth.getAccessToken();
    if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
    }

    const options = { method, headers };

    if (body) {
        if (isFormData) {
            options.body = body;
        } else {
            headers['Content-Type'] = 'application/json';
            options.body = JSON.stringify(body);
        }
    }

    const response = await fetch(`${API_BASE_URL}/api${endpoint}`, options);

    if (response.status === 401) {
        // Token expired or invalid, log out the user
        Auth.logout();
        history.pushState(null, null, '/login');
        router();
        throw new Error('Session expired. Please log in again.');
    }

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'An unknown error occurred' }));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }

    if (response.status === 204) {
        return null; // No content
    }

    return response.json();
};

export const logAnalyticsEvent = async (projectId, eventType) => {
    try {
        await apiRequest('/analytics', 'POST', { project_id: projectId, event_type: eventType });
        console.log(`Logged event: ${eventType} for project ${projectId}`);
    } catch (error) {
        console.error(`Failed to log analytics event: ${error.message}`);
    }
}
