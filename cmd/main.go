package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/forkspacer/api-server/pkg/api"
	apiv1 "github.com/forkspacer/api-server/pkg/api/v1"
	"github.com/forkspacer/api-server/pkg/config"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenForTermination(func() { cancel() })

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	apiConfig, errs := config.NewAPIConfig()
	if errs != nil {
		for _, err := range errs.Errors {
			logger.Error("Config error", zap.Error(err))
		}
		return
	}

	// Switch to production logger if not in development mode.
	if !apiConfig.Dev {
		logger, err = zap.NewProduction()
		if err != nil {
			panic(err)
		}
	}

	forkspacerWorkspaceService, err := forkspacer.NewForkspacerWorkspaceService()
	if err != nil {
		logger.Fatal("Failed to create Forkspacer workspace service", zap.Error(err))
	}

	forkspacerModuleService, err := forkspacer.NewForkspacerModuleService()
	if err != nil {
		logger.Fatal("Failed to create Forkspacer module service", zap.Error(err))
	}

	logger.Info("Starting API server", zap.Uint16("port", apiConfig.APIPort))

	if err := api.Run(ctx,
		apiConfig.APIPort,
		apiv1.NewRouter(logger, forkspacerWorkspaceService, forkspacerModuleService),
	); err != nil {
		logger.Error("API server failed to run", zap.Error(err), zap.Uint16("port", apiConfig.APIPort))
	}

	logger.Info("API server stopped", zap.Uint16("port", apiConfig.APIPort))
}

func listenForTermination(do func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	do()
}
