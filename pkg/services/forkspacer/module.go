package forkspacer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/forkspacer/api-server/pkg/utils"
	batchv1 "github.com/forkspacer/forkspacer/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ForkspacerModuleService struct {
	client client.Client
}

func NewForkspacerModuleService() (*ForkspacerModuleService, error) {
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

	return &ForkspacerModuleService{client: ctrlClient}, nil
}

type ModuleSource struct {
	Raw     map[string]any
	HttpURL *string
}

type ModuleCreateIn struct {
	Name       string
	Namespace  *string
	Workspace  ResourceReference
	Source     ModuleSource
	Config     map[string]any
	Hibernated bool
}

func (s ForkspacerModuleService) Create(ctx context.Context, moduleIn ModuleCreateIn) (*batchv1.Module, error) {
	sourceRaw, err := json.Marshal(moduleIn.Source.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal source.raw: %w", err)
	}

	config, err := json.Marshal(moduleIn.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if moduleIn.Namespace == nil {
		moduleIn.Namespace = utils.ToPtr("default")
	}

	module := &batchv1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      moduleIn.Name,
			Namespace: *moduleIn.Namespace,
		},
		Spec: batchv1.ModuleSpec{
			Workspace: batchv1.ModuleWorkspaceReference{
				Name:      moduleIn.Workspace.Name,
				Namespace: moduleIn.Workspace.Namespace,
			},
			Source: batchv1.ModuleSource{
				Raw: &runtime.RawExtension{
					Raw: sourceRaw,
				},
				HttpURL: moduleIn.Source.HttpURL,
			},
			Config: &runtime.RawExtension{
				Raw: config,
			},
			Hibernated: utils.ToPtr(moduleIn.Hibernated),
		},
	}

	return module, s.client.Create(ctx, module)
}
