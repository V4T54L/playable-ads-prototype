const Auth = {
    saveTokens(accessToken, refreshToken) {
        localStorage.setItem('accessToken', accessToken);
        if (refreshToken) {
            localStorage.setItem('refreshToken', refreshToken);
        }
    },

    getAccessToken() {
        return localStorage.getItem('accessToken');
    },
    
    getRefreshToken() {
        return localStorage.getItem('refreshToken');
    },

    logout() {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
    },
    
    isAuthenticated() {
        return !!this.getAccessToken();
    }
};

export default Auth;
