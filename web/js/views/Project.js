import { apiRequest, logAnalyticsEvent } from '../api.js';

export default class {
    constructor() {
        const urlParams = new URLSearchParams(window.location.search);
        this.projectId = urlParams.get('id');
        document.title = `Project Details`;
    }

    async getProjectDetails() {
        try {
            const project = await apiRequest(`/projects/${this.projectId}`);
            const assets = await apiRequest(`/projects/${this.projectId}/assets`);

            // Log impression event
            logAnalyticsEvent(this.projectId, 'impression');

            return { project, assets };
        } catch (error) {
            return { error: error.message };
        }
    }

    renderAssets(assets) {
        if (!assets || assets.length === 0) {
            return '<p class="text-center text-gray-500 py-4">No assets found. Upload one below!</p>';
        }
        return `
            <div class="overflow-x-auto">
                <table class="min-w-full bg-white border border-gray-200">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Filename</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Size (KB)</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Uploaded</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200">
                        ${assets.map(a => `
                            <tr class="hover:bg-gray-50">
                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${a.original_filename}</td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${a.mime_type}</td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${(a.size / 1024).toFixed(2)}</td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${new Date(a.created_at).toLocaleString()}</td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                    <button class="bg-green-500 hover:bg-green-600 text-white font-bold py-1 px-3 rounded-lg text-xs" data-project-id="${this.projectId}" data-asset-id="${a.id}">Render</button>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }

    async render() {
        if (!this.projectId) {
            return `<p class="text-red-500 text-center">No project ID provided. <a href="/dashboard" class="text-blue-500 underline" data-link>Go to Dashboard</a></p>`;
        }

        const { project, assets, error } = await this.getProjectDetails();

        if (error) {
            return `<p class="text-red-500 text-center">Failed to load project: ${error}</p>`;
        }

        document.title = `Project: ${project.title}`;

        return `
            <div class="space-y-12">
                <section id="project-details">
                    <h2 class="text-3xl font-bold text-gray-800">${project.title}</h2>
                    <p class="text-gray-600 mt-2">${project.description}</p>
                </section>

                <section>
                    <h3 class="text-2xl font-bold text-gray-800 mb-4">Assets</h3>
                    <div id="assets-list" class="bg-gray-50 p-4 rounded-lg border border-gray-200">
                        ${this.renderAssets(assets)}
                    </div>
                </section>

                <section>
                    <h3 class="text-2xl font-bold text-gray-800 mb-4">Upload New Asset</h3>
                    <form id="upload-asset-form" class="bg-gray-50 p-6 rounded-lg border border-gray-200">
                        <div class="mb-4">
                            <label for="asset-file" class="block text-sm font-medium text-gray-700 mb-1">Select File (Max 10MB)</label>
                            <input type="file" id="asset-file" name="asset" required accept="video/mp4" class="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100">
                        </div>
                        <button type="submit" class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg transition-colors">Upload Asset</button>
                        <p id="upload-asset-error" class="text-red-500 text-center mt-2 h-4"></p>
                    </form>
                </section>
            </div>
        `;
    }

    async after_render() {
        if (!this.projectId) return;

        // --- Handle Asset Upload ---
        const uploadForm = document.getElementById('upload-asset-form');
        uploadForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const fileInput = document.getElementById('asset-file');
            const errorElement = document.getElementById('upload-asset-error');
            const uploadBtn = uploadForm.querySelector('button');

            if (fileInput.files.length === 0) {
                errorElement.textContent = 'Please select a file.';
                return;
            }

            const formData = new FormData();
            formData.append('file', fileInput.files[0]);

            uploadBtn.textContent = 'Uploading...';
            uploadBtn.disabled = true;
            errorElement.textContent = '';

            try {
                await apiRequest(`/projects/${this.projectId}/assets`, 'POST', formData, true);
                uploadForm.reset();
                // Refresh assets list
                const { assets } = await this.getProjectDetails();
                document.getElementById('assets-list').innerHTML = this.renderAssets(assets);
            } catch (error) {
                errorElement.textContent = `Upload failed: ${error.message}`;
            } finally {
                uploadBtn.textContent = 'Upload Asset';
                uploadBtn.disabled = false;
            }
        });

        // --- Handle Render Button Clicks ---
        const assetsList = document.getElementById('assets-list');
        assetsList.addEventListener('click', async (e) => {
            if (e.target.tagName === 'BUTTON' && e.target.dataset.assetId) {
                const button = e.target;
                const projectId = button.dataset.projectId;
                const assetId = button.dataset.assetId;

                button.textContent = 'Starting...';
                button.disabled = true;

                try {
                    await apiRequest(`/projects/${projectId}/render/${assetId}`, 'POST');
                    alert('Render job enqueued! Check the Jobs page.');
                    // Log click event
                    logAnalyticsEvent(projectId, 'click');
                } catch (error) {
                    alert(`Failed to start render job: ${error.message}`);
                } finally {
                    button.textContent = 'Render';
                    button.disabled = false;
                }
            }
        });
    }
}
