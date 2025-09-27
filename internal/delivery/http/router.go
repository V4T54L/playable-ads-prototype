package http

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "backend/docs"
	"backend/internal/delivery/http/handler"
	http_middleware "backend/internal/delivery/http/middleware"
	"backend/pkg/auth"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	projectHandler *handler.ProjectHandler,
	assetHandler *handler.AssetHandler,
	jobHandler *handler.JobHandler,
	analyticsHandler *handler.AnalyticsHandler,
	jwtMaker *auth.JWTMaker,
	logger *slog.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(http_middleware.NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	r.Get("/analytics", analyticsHandler.ListEvents) // Public endpoint to view all events

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(http_middleware.AuthMiddleware(jwtMaker))

		// Project routes
		r.Route("/projects", func(r chi.Router) {
			r.Post("/", projectHandler.CreateProject)
			r.Get("/", projectHandler.ListProjects)
			r.Get("/{id}", projectHandler.GetProject)

			// Asset routes nested under projects
			r.Post("/{id}/assets", assetHandler.UploadAsset)
			r.Get("/{id}/assets", assetHandler.ListAssets)

			// Render routes
			r.Post("/{projectId}/render/{assetId}", jobHandler.EnqueueRenderJob)
		})

		// Asset retrieval
		r.Get("/assets/{id}", assetHandler.GetAsset)

		// Job routes
		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", jobHandler.ListJobs)
			r.Get("/{id}", jobHandler.GetJobStatus)
		})

		// Outputs route
		r.Get("/outputs", jobHandler.ListOutputs)

		// Analytics logging
		r.Post("/analytics", analyticsHandler.LogEvent)
	})

	// Serve static files for the frontend
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	FileServer(r, "/", filesDir)

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		// Check if the file exists
		f, err := root.Open(r.URL.Path)
		if os.IsNotExist(err) {
			// If not found, serve index.html for SPA routing
			// Ensure root is of type http.Dir
			if dir, ok := root.(http.Dir); ok {
				http.ServeFile(w, r, filepath.Join(string(dir), "web", "index.html"))
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		fs.ServeHTTP(w, r)
	})
}
