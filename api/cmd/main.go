package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/forkspacer/api-server/pkg/api"
	apiv1 "github.com/forkspacer/api-server/pkg/api/v1"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenForTermination(func() { cancel() })

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger.Info("Starting API server")
	if err := api.Run(ctx, apiv1.NewRouter(logger)); err != nil {
		logger.Error("API server failed to run", zap.Error(err))
	}
	logger.Info("API server stopped")
}

func listenForTermination(do func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	do()
}
