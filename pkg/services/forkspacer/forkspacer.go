package forkspacer

import (
	"context"
	"fmt"

	batchv1 "github.com/forkspacer/forkspacer/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ForkspacerService struct {
	client client.Client
}

func NewForkspacerService() (*ForkspacerService, error) {
	restConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := batchv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add batch.environment.sh/v1 to scheme: %w", err)
	}

	ctrlClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	return &ForkspacerService{client: ctrlClient}, nil
}

func (s ForkspacerService) Create(ctx context.Context) error {
	return s.client.Create(ctx, &batchv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		Spec: batchv1.WorkspaceSpec{
			Type: batchv1.WorkspaceTypeKubernetes,
			Connection: &batchv1.WorkspaceConnection{
				Type: batchv1.WorkspaceConnectionTypeInCluster,
			},
		},
	})
}
