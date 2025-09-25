package v1

import (
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/v1/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger) http.Handler {
	workspaceHandler := handlers.NewWorkspaceHandler(logger)

	apiRouter := chi.NewRouter()
	apiRouter.Route("/workspace", func(r chi.Router) {
		r.Get("/", workspaceHandler.Create)
		r.Patch("/", workspaceHandler.Update)
		r.Delete("/", workspaceHandler.Delete)
	})

	baseRouter := chi.NewRouter()
	baseRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	baseRouter.Mount("/v1", apiRouter)

	return baseRouter
}
