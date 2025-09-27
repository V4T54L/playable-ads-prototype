import { apiRequest } from '../api.js';

export default class {
    constructor() {
        document.title = 'Analytics';
    }

    async render() {
        let content = '<p class="text-center text-gray-500 py-4">Loading analytics...</p>';
        try {
            const events = await apiRequest('/analytics');
            if (!events || events.length === 0) {
                content = '<p class="text-center text-gray-500 py-4">No analytics events found.</p>';
            } else {
                content = `
                    <div class="overflow-x-auto">
                        <table class="min-w-full bg-white border border-gray-200">
                            <thead class="bg-gray-50">
                                <tr>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Event ID</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Project ID</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User ID</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Event Type</th>
                                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-gray-200">
                                ${events.map(e => `
                                    <tr class="hover:bg-gray-50">
                                        <td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">${e.id}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">${e.project_id}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">${e.user_id}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${e.event_type}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${new Date(e.created_at).toLocaleString()}</td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    </div>
                `;
            }
        } catch (error) {
            content = `<p class="text-red-500 text-center">Failed to load analytics: ${error.message}</p>`;
        }
        
        return `
            <h2 class="text-3xl font-bold text-gray-800 mb-2">Analytics Events</h2>
            <p class="text-gray-500 mb-6">This is a public list of all analytics events logged in the system.</p>
            <div id="analytics-list" class="bg-gray-50 p-4 rounded-lg border border-gray-200">
                ${content}
            </div>
        `;
    }

    async after_render() {
        // No specific event listeners needed for this static view
    }
}
