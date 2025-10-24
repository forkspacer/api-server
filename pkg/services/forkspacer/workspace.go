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
	"k8s.io/client-go/util/retry"
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
		return nil, fmt.Errorf("failed to add batch.forkspacer.com/v1 to scheme: %w", err)
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

func (s ForkspacerWorkspaceService) ListKubeconfigSecrets(
	ctx context.Context,
	limit int64, continueToken *string,
) (*corev1.SecretList, error) {
	options := []client.ListOption{
		client.MatchingLabels{BaseLabel: Labels.WorkspaceKubeconfigSecret},
		client.Limit(limit),
	}

	if continueToken != nil {
		options = append(options, client.Continue(*continueToken))
	}

	secrets := &corev1.SecretList{}
	err := s.client.List(ctx, secrets, options...)

	return secrets, err
}

type WorkspaceCreateConnectionIn struct {
	Type   string
	Secret *ResourceReference
	Key    *string
}

type WorkspaceAutoHibernationIn struct {
	Enabled      bool
	Schedule     string
	WakeSchedule *string
}

type ManagedClusterIn struct {
	Backend *string
	Distro  *string
}

type WorkspaceCreateIn struct {
	Name            string
	Namespace       *string
	Type            *string
	From            *ResourceReference
	Hibernated      bool
	Connection      *WorkspaceCreateConnectionIn
	ManagedCluster  *ManagedClusterIn
	AutoHibernation *WorkspaceAutoHibernationIn
}

func (s ForkspacerWorkspaceService) Create(
	ctx context.Context, workspaceIn WorkspaceCreateIn,
) (*batchv1.Workspace, error) {
	if workspaceIn.Namespace == nil {
		workspaceIn.Namespace = utils.ToPtr("default")
	}

	// Set default workspace type to kubernetes if not specified
	workspaceType := batchv1.WorkspaceTypeKubernetes
	if workspaceIn.Type != nil {
		workspaceType = batchv1.WorkspaceType(*workspaceIn.Type)
	}

	workspace := &batchv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspaceIn.Name,
			Namespace: *workspaceIn.Namespace,
		},
		Spec: batchv1.WorkspaceSpec{
			Type:       workspaceType,
			Hibernated: workspaceIn.Hibernated,
			Connection: batchv1.WorkspaceConnection{
				Type: batchv1.WorkspaceConnectionType(workspaceIn.Connection.Type),
			},
		},
	}

	if workspaceIn.From != nil {
		workspace.Spec.From = &batchv1.WorkspaceFromReference{
			Name:      workspaceIn.From.Name,
			Namespace: workspaceIn.From.Namespace,
		}
	}

	if workspaceIn.Connection != nil && workspaceIn.Connection.Secret != nil {
		workspace.Spec.Connection.SecretReference = &batchv1.WorkspaceConnectionSecretReference{
			Name:      workspaceIn.Connection.Secret.Name,
			Namespace: workspaceIn.Connection.Secret.Namespace,
		}
		if workspaceIn.Connection.Key != nil {
			workspace.Spec.Connection.SecretReference.Key = *workspaceIn.Connection.Key
		}
	}

	if workspaceIn.AutoHibernation != nil {
		workspace.Spec.AutoHibernation.Enabled = workspaceIn.AutoHibernation.Enabled
		workspace.Spec.AutoHibernation.Schedule = workspaceIn.AutoHibernation.Schedule
		workspace.Spec.AutoHibernation.WakeSchedule = workspaceIn.AutoHibernation.WakeSchedule
	}

	if workspaceIn.ManagedCluster != nil {
		workspace.Spec.ManagedCluster = &batchv1.ManagedClusterSpec{}
		if workspaceIn.ManagedCluster.Backend != nil {
			workspace.Spec.ManagedCluster.Backend = batchv1.ManagedClusterBackend(*workspaceIn.ManagedCluster.Backend)
		}
		if workspaceIn.ManagedCluster.Distro != nil {
			workspace.Spec.ManagedCluster.Distro = *workspaceIn.ManagedCluster.Distro
		}
	}

	return workspace, s.client.Create(ctx, workspace)
}

func (s ForkspacerWorkspaceService) Delete(ctx context.Context, name string, namespace *string) error {
	if namespace == nil {
		namespace = utils.ToPtr("default")
	}

	workspace := &batchv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: *namespace,
		},
	}

	return s.client.Delete(ctx, workspace)
}

func (s ForkspacerWorkspaceService) List(
	ctx context.Context,
	limit int64,
	continueToken *string,
) (*batchv1.WorkspaceList, error) {
	options := []client.ListOption{
		client.Limit(limit),
	}

	if continueToken != nil {
		options = append(options, client.Continue(*continueToken))
	}

	workspaces := &batchv1.WorkspaceList{}
	err := s.client.List(ctx, workspaces, options...)

	return workspaces, err
}

type WorkspaceUpdateIn struct {
	Name            string
	Namespace       *string
	Hibernated      *bool
	AutoHibernation *WorkspaceAutoHibernationIn
	ManagedCluster  *ManagedClusterIn
}

func (s ForkspacerWorkspaceService) Update(
	ctx context.Context,
	updateIn WorkspaceUpdateIn,
) (*batchv1.Workspace, error) {
	if updateIn.Namespace == nil {
		updateIn.Namespace = utils.ToPtr("default")
	}

	workspace := &batchv1.Workspace{}

	return workspace, retry.RetryOnConflict(
		retry.DefaultRetry,
		func() error {

			if err := s.client.Get(ctx, client.ObjectKey{
				Name:      updateIn.Name,
				Namespace: *updateIn.Namespace,
			}, workspace); err != nil {
				return err
			}

			// Update only the allowed fields
			if updateIn.Hibernated != nil {
				workspace.Spec.Hibernated = *updateIn.Hibernated
			}

			if updateIn.AutoHibernation != nil {
				workspace.Spec.AutoHibernation.Enabled = updateIn.AutoHibernation.Enabled
				workspace.Spec.AutoHibernation.Schedule = updateIn.AutoHibernation.Schedule
				workspace.Spec.AutoHibernation.WakeSchedule = updateIn.AutoHibernation.WakeSchedule
			}

			if updateIn.ManagedCluster != nil {
				if workspace.Spec.ManagedCluster == nil {
					workspace.Spec.ManagedCluster = &batchv1.ManagedClusterSpec{}
				}
				if updateIn.ManagedCluster.Backend != nil {
					workspace.Spec.ManagedCluster.Backend = batchv1.ManagedClusterBackend(*updateIn.ManagedCluster.Backend)
				}
				if updateIn.ManagedCluster.Distro != nil {
					workspace.Spec.ManagedCluster.Distro = *updateIn.ManagedCluster.Distro
				}
			}

			return s.client.Update(ctx, workspace)
		},
	)
}
