import { apiRequest } from '../api.js';

export default class {
    constructor() {
        document.title = 'Dashboard';
    }

    async render() {
        const projects = await apiRequest('/projects');
        
        let projectsGrid = '<p class="text-center text-gray-500 col-span-full">No projects found. Create one above!</p>';
        if (projects && projects.length > 0) {
            projectsGrid = projects.map(p => `
                <a href="/project?id=${p.id}" class="block bg-white p-6 rounded-lg border border-gray-200 hover:shadow-md hover:-translate-y-1 transition-all" data-link>
                    <h3 class="text-xl font-semibold text-gray-800 truncate">${p.title}</h3>
                    <p class="text-gray-600 mt-2">${p.description || 'No description'}</p>
                </a>
            `).join('');
        }

        return `
            <div class="space-y-12">
                <section>
                    <h2 class="text-2xl font-bold text-gray-800 mb-4">Create New Project</h2>
                    <form id="create-project-form" class="bg-gray-50 p-6 rounded-lg border border-gray-200">
                        <div class="mb-4">
                            <label for="title" class="block text-sm font-medium text-gray-700 mb-1">Project Title</label>
                            <input type="text" id="title" name="title" required minlength="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                        </div>
                        <div class="mb-4">
                            <label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
                            <textarea id="description" name="description" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"></textarea>
                        </div>
                        <button type="submit" class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg transition-colors">Create Project</button>
                        <p id="create-project-error" class="text-red-500 text-center mt-2 h-4"></p>
                    </form>
                </section>
                
                <section>
                    <h2 class="text-2xl font-bold text-gray-800 mb-4">My Projects</h2>
                    <div id="projects-list" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        ${projectsGrid}
                    </div>
                </section>
            </div>
        `;
    }

    async after_render() {
        const form = document.getElementById('create-project-form');
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const title = form.title.value;
            const description = form.description.value;
            const errorElement = document.getElementById('create-project-error');

            try {
                await apiRequest('/projects', 'POST', { title, description });
                form.reset();
                errorElement.textContent = '';
                
                // Re-render the projects list without full page reload
                const projects = await apiRequest('/projects');
                const projectsList = document.getElementById('projects-list');
                if (projects && projects.length > 0) {
                     projectsList.innerHTML = projects.map(p => `
                        <a href="/project?id=${p.id}" class="block bg-white p-6 rounded-lg border border-gray-200 hover:shadow-md hover:-translate-y-1 transition-all" data-link>
                            <h3 class="text-xl font-semibold text-gray-800 truncate">${p.title}</h3>
                            <p class="text-gray-600 mt-2">${p.description || 'No description'}</p>
                        </a>
                    `).join('');
                } else {
                    projectsList.innerHTML = '<p class="text-center text-gray-500 col-span-full">No projects found.</p>';
                }

            } catch (error) {
                errorElement.textContent = error.message;
            }
        });
    }
}
