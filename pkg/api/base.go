package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func Run(ctx context.Context, port uint16, routers ...http.Handler) error {
	baseRouter := chi.NewRouter()

	baseRouter.Use(middleware.Recoverer)
	for _, router := range routers {
		baseRouter.Mount("/api", router)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: baseRouter,
	}

	listenerErrChan := make(chan error)
	go func() {
		listenerErrChan <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-listenerErrChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("error while serving http: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("error while shutting down http server: %v", err)
		}
	}

	return nil
}
