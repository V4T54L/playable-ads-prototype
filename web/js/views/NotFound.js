export default class {
    constructor() {
        document.title = '404 - Not Found';
    }

    async render() {
        return `
            <div class="min-h-screen flex items-center justify-center text-center">
                <div>
                    <h1 class="text-6xl font-bold text-blue-600">404</h1>
                    <p class="text-2xl mt-4 text-gray-800">Page Not Found</p>
                    <p class="text-gray-500 mt-2">The page you are looking for does not exist.</p>
                    <a href="/" class="mt-6 inline-block bg-blue-600 text-white font-bold py-2 px-4 rounded-lg hover:bg-blue-700 transition-colors" data-link>
                        Go to Dashboard
                    </a>
                </div>
            </div>
        `;
    }

    async after_render() {}
}
