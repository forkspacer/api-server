package v1

import (
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/v1/handlers"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

func NewRouter(
	logger *zap.Logger,
	forkspacerWorkspaceService *forkspacer.ForkspacerWorkspaceService,
	forkspacerModuleService *forkspacer.ForkspacerModuleService,
) http.Handler {
	workspaceHandler := handlers.NewWorkspaceHandler(logger, forkspacerWorkspaceService)
	moduleHandler := handlers.NewModuleHandler(logger, forkspacerModuleService)

	apiRouter := chi.NewRouter()
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
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	baseRouter.Mount("/v1", apiRouter)

	return baseRouter
}
