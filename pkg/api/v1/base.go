package v1

import (
	_ "embed"
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/v1/handlers"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

//go:embed docs.html
var docsHTML []byte

//go:embed openapi.yaml
var openAPISpec []byte

func NewRouter(
	logger *zap.Logger,
	forkspacerWorkspaceService *forkspacer.ForkspacerWorkspaceService,
	forkspacerModuleService *forkspacer.ForkspacerModuleService,
) http.Handler {
	workspaceHandler := handlers.NewWorkspaceHandler(logger, forkspacerWorkspaceService)
	moduleHandler := handlers.NewModuleHandler(logger, forkspacerModuleService)

	apiRouter := chi.NewRouter()

	apiRouter.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err := w.Write(docsHTML)
		if err != nil {
			logger.Error("failed to write docs HTML response", zap.Error(err))
		}
	})

	// Serve OpenAPI spec
	apiRouter.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
		_, err := w.Write(openAPISpec)
		if err != nil {
			logger.Error("failed to write OpenAPI spec response", zap.Error(err))
		}
	})

	apiRouter.Route("/workspace", func(r chi.Router) {
		r.Post("/", workspaceHandler.CreateHandle)
		r.Patch("/", workspaceHandler.UpdateHandle)
		r.Delete("/", workspaceHandler.DeleteHandle)
		r.Get("/list", workspaceHandler.ListHandle)

		r.Route("/connection", func(r chi.Router) {
			r.Route("/kubeconfig", func(r chi.Router) {
				r.Post("/", workspaceHandler.CreateKubeconfigSecretHandle)
				r.Delete("/", workspaceHandler.DeleteKubeconfigSecretHandle)
				r.Get("/list", workspaceHandler.ListKubeconfigSecretsHandle)
			})
		})
	})

	apiRouter.Route("/module", func(r chi.Router) {
		r.Post("/", moduleHandler.CreateHandle)
		r.Patch("/", moduleHandler.UpdateHandle)
		r.Delete("/", moduleHandler.DeleteHandle)
		r.Get("/list", moduleHandler.ListHandle)
	})

	baseRouter := chi.NewRouter()
	baseRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))

	baseRouter.Mount("/v1", apiRouter)

	return baseRouter
}
