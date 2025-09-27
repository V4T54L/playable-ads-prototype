// This component wraps our authenticated pages, providing a consistent layout.
const Layout = (pageContent) => {
    return `
        <div class="min-h-screen flex items-start justify-center p-4 sm:p-6 lg:p-8">
            <div class="container max-w-5xl w-full bg-white rounded-xl shadow-lg">
                <nav class="bg-gray-50 p-4 border-b border-gray-200 flex justify-between items-center rounded-t-xl">
                    <div class="flex items-center space-x-6">
                        <a href="/dashboard" class="text-gray-600 hover:text-blue-600 font-medium" data-link>Dashboard</a>
                        <a href="/jobs" class="text-gray-600 hover:text-blue-600 font-medium" data-link>Jobs</a>
                        <a href="/analytics" class="text-gray-600 hover:text-blue-600 font-medium" data-link>Analytics</a>
                    </div>
                    <button id="logout-button" class="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded-lg transition-colors">
                        Logout
                    </button>
                </nav>
                <main class="p-6 md:p-8">
                    ${pageContent}
                </main>
            </div>
        </div>
    `;
};

export default Layout;
