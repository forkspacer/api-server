package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func Run(ctx context.Context, router http.Handler) error {
	httpServer := &http.Server{
		Addr:    ":8421",
		Handler: router,
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
