import { apiRequest } from '../api.js';

export default class {
    constructor() {
        document.title = 'Rendering Jobs';
        this.intervalId = null; // To hold the interval ID for cleanup
    }
    
    renderJobStatus(status) {
        const lowerStatus = status.toLowerCase();
        const colors = {
            pending: 'text-yellow-600 bg-yellow-100',
            processing: 'text-blue-600 bg-blue-100',
            done: 'text-green-600 bg-green-100',
            failed: 'text-red-600 bg-red-100',
        };
        return `<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${colors[lowerStatus] || 'text-gray-600 bg-gray-100'}">${status}</span>`;
    }

    async loadAndRenderJobs() {
        const listElement = document.getElementById('jobs-list-container');
        if (!listElement) return;

        try {
            const jobs = await apiRequest('/jobs');
            if (!jobs || jobs.length === 0) {
                listElement.innerHTML = '<p class="text-center text-gray-500 py-4">No jobs found.</p>';
                return;
            }
            listElement.innerHTML = `
                <div class="overflow-x-auto">
                    <table class="min-w-full bg-white border border-gray-200">
                        <thead class="bg-gray-50">
                            <tr>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Job ID</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Output / Error</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Completed</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-200">
                            ${jobs.map(j => `
                                <tr class="hover:bg-gray-50">
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">${j.id}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm">${this.renderJobStatus(j.status)}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${j.status === 'done' ? `<a href="download/${j.output_path}" target="_blank" class="text-blue-600 hover:underline">Download</a>` : (j.error_message || 'N/A')}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${new Date(j.created_at).toLocaleString()}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${j.completed_at ? new Date(j.completed_at).toLocaleString() : 'N/A'}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            `;
        } catch (error) {
            listElement.innerHTML = `<p class="text-red-500 text-center">Failed to load jobs: ${error.message}</p>`;
        }
    }

    async render() {
        return `
            <h2 class="text-3xl font-bold text-gray-800 mb-2">My Rendering Jobs</h2>
            <p class="text-gray-500 mb-6">This page automatically refreshes every 5 seconds.</p>
            <div id="jobs-list-container" class="bg-gray-50 p-4 rounded-lg border border-gray-200">
                <p class="text-center text-gray-500 py-4">Loading jobs...</p>
            </div>
        `;
    }

    async after_render() {
        // Initial load
        this.loadAndRenderJobs();
        
        // Start polling every 5 seconds
        this.intervalId = setInterval(() => this.loadAndRenderJobs(), 5000);

        // **IMPORTANT**: Clean up the interval when the view is destroyed.
        // This is a simplified cleanup; a real router would have a `destroy` or `unmount` method.
        // For now, we rely on the fact that creating a new view will overwrite the old one,
        // but we'll manually clear the interval if we navigate away.
        const cleanup = (e) => {
            if (e.target.closest('[data-link]') && window.location.pathname !== '/jobs') {
                 clearInterval(this.intervalId);
                 document.body.removeEventListener('click', cleanup);
            }
        };
        document.body.addEventListener('click', cleanup);
        window.addEventListener('popstate', () => clearInterval(this.intervalId), { once: true });
    }
}
