package forkspacer

import (
	"context"
	"fmt"

	"github.com/forkspacer/api-server/pkg/utils"
	batchv1 "github.com/forkspacer/forkspacer/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ForkspacerWorkspaceService struct {
	client client.Client
}

func NewForkspacerWorkspaceService() (*ForkspacerWorkspaceService, error) {
	restConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add go client to schemes: %w", err)
	}
	if err := batchv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add batch.environment.sh/v1 to scheme: %w", err)
	}

	ctrlClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	return &ForkspacerWorkspaceService{client: ctrlClient}, nil
}

func (s ForkspacerWorkspaceService) CreateKubeconfigSecret(
	ctx context.Context,
	name string, namespace *string,
	kubeconfigData []byte,
) (*corev1.Secret, error) {
	if namespace == nil {
		namespace = utils.ToPtr("default")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: *namespace,
			Labels: map[string]string{
				BaseLabel: Labels.WorkspaceKubeconfigSecret,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"kubeconfig": kubeconfigData,
		},
	}

	return secret, s.client.Create(ctx, secret)
}

func (s ForkspacerWorkspaceService) DeleteKubeconfigSecret(
	ctx context.Context,
	name string, namespace *string,
) error {
	if namespace == nil {
		namespace = utils.ToPtr("default")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: *namespace,
			Labels: map[string]string{
				BaseLabel: Labels.WorkspaceKubeconfigSecret,
			},
		},
	}

	return s.client.Delete(ctx, secret)
}
