package forkspacer

import (
	"fmt"

	"github.com/forkspacer/api-server/pkg/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ForkspacerService struct {
	kubernetesClient *kubernetes.Clientset
}

func NewForkspacerService(kubernetesConfig config.KubernetesConfig) (*ForkspacerService, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubernetesConfig.KubeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes config: %w", err)
	}

	kubernetesClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	return &ForkspacerService{kubernetesClient}, nil
}
