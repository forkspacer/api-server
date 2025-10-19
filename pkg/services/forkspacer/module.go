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
	"sigs.k8s.io/yaml"
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
		return nil, fmt.Errorf("failed to add batch.forkspacer.com/v1 to scheme: %w", err)
	}

	ctrlClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	return &ForkspacerModuleService{client: ctrlClient}, nil
}

type ModuleSourceConfigMapRef struct {
	Name      string
	Namespace string
	Key       *string
}

type ModuleSourceChartRepository struct {
	URL     string
	Chart   string
	Version *string
}

type ModuleSourceChartGit struct {
	Repo     string
	Path     string
	Revision string
}

type ModuleSourceChartRef struct {
	ConfigMap  *ModuleSourceConfigMapRef
	Repository *ModuleSourceChartRepository
	Git        *ModuleSourceChartGit
}

type ModuleSourceExistingHelmReleaseRef struct {
	Name        string
	Namespace   string
	ChartSource ModuleSourceChartRef
	Values      map[string]any
}

type ModuleSource struct {
	Raw                 []byte
	HttpURL             *string
	ConfigMap           *ModuleSourceConfigMapRef
	ExistingHelmRelease *ModuleSourceExistingHelmReleaseRef
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
	config, err := json.Marshal(moduleIn.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if moduleIn.Namespace == nil {
		moduleIn.Namespace = utils.ToPtr("default")
	}

	// Convert YAML bytes to JSON for RawExtension
	var sourceRawJSON []byte
	if moduleIn.Source.Raw != nil {
		sourceRawJSON, err = yaml.YAMLToJSON(moduleIn.Source.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to convert YAML to JSON for source.raw: %w", err)
		}
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
			Source: batchv1.ModuleSource{},
			Config: &runtime.RawExtension{
				Raw: config,
			},
			Hibernated: moduleIn.Hibernated,
		},
	}

	// Set source fields
	if moduleIn.Source.Raw != nil {
		module.Spec.Source.Raw = &runtime.RawExtension{
			Raw: sourceRawJSON,
		}
	}

	if moduleIn.Source.HttpURL != nil {
		module.Spec.Source.HttpURL = moduleIn.Source.HttpURL
	}

	if moduleIn.Source.ConfigMap != nil {
		module.Spec.Source.ConfigMap = &batchv1.ModuleSourceConfigMapRef{
			Name:      moduleIn.Source.ConfigMap.Name,
			Namespace: moduleIn.Source.ConfigMap.Namespace,
		}
		if moduleIn.Source.ConfigMap.Key != nil {
			module.Spec.Source.ConfigMap.Key = *moduleIn.Source.ConfigMap.Key
		}
	}

	if moduleIn.Source.ExistingHelmRelease != nil {
		valuesJSON, err := json.Marshal(moduleIn.Source.ExistingHelmRelease.Values)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal existingHelmRelease values: %w", err)
		}

		module.Spec.Source.ExistingHelmRelease = &batchv1.ModuleSourceExistingHelmReleaseRef{
			Name:        moduleIn.Source.ExistingHelmRelease.Name,
			Namespace:   moduleIn.Source.ExistingHelmRelease.Namespace,
			ChartSource: batchv1.ModuleSourceChartRef{},
		}

		if moduleIn.Source.ExistingHelmRelease.Values != nil {
			module.Spec.Source.ExistingHelmRelease.Values = &runtime.RawExtension{
				Raw: valuesJSON,
			}
		}

		// Set chart source
		if moduleIn.Source.ExistingHelmRelease.ChartSource.ConfigMap != nil {
			module.Spec.Source.ExistingHelmRelease.ChartSource.ConfigMap = &batchv1.ModuleSourceConfigMapRef{
				Name:      moduleIn.Source.ExistingHelmRelease.ChartSource.ConfigMap.Name,
				Namespace: moduleIn.Source.ExistingHelmRelease.ChartSource.ConfigMap.Namespace,
			}
			if moduleIn.Source.ExistingHelmRelease.ChartSource.ConfigMap.Key != nil {
				module.Spec.Source.ExistingHelmRelease.ChartSource.ConfigMap.Key = *moduleIn.Source.ExistingHelmRelease.ChartSource.ConfigMap.Key
			}
		}

		if moduleIn.Source.ExistingHelmRelease.ChartSource.Repository != nil {
			module.Spec.Source.ExistingHelmRelease.ChartSource.Repository = &batchv1.ModuleSourceChartRepository{
				URL:   moduleIn.Source.ExistingHelmRelease.ChartSource.Repository.URL,
				Chart: moduleIn.Source.ExistingHelmRelease.ChartSource.Repository.Chart,
			}
			if moduleIn.Source.ExistingHelmRelease.ChartSource.Repository.Version != nil {
				module.Spec.Source.ExistingHelmRelease.ChartSource.Repository.Version = moduleIn.Source.ExistingHelmRelease.ChartSource.Repository.Version
			}
		}

		if moduleIn.Source.ExistingHelmRelease.ChartSource.Git != nil {
			module.Spec.Source.ExistingHelmRelease.ChartSource.Git = &batchv1.ModuleSourceChartGit{
				Repo:     moduleIn.Source.ExistingHelmRelease.ChartSource.Git.Repo,
				Path:     moduleIn.Source.ExistingHelmRelease.ChartSource.Git.Path,
				Revision: moduleIn.Source.ExistingHelmRelease.ChartSource.Git.Revision,
			}
		}
	}

	return module, s.client.Create(ctx, module)
}

type ModuleUpdateIn struct {
	Name       string
	Namespace  *string
	Hibernated *bool
}

func (s ForkspacerModuleService) Update(
	ctx context.Context,
	updateIn ModuleUpdateIn,
) (*batchv1.Module, error) {
	if updateIn.Namespace == nil {
		updateIn.Namespace = utils.ToPtr("default")
	}

	module := &batchv1.Module{}

	if err := s.client.Get(ctx, client.ObjectKey{
		Name:      updateIn.Name,
		Namespace: *updateIn.Namespace,
	}, module); err != nil {
		return nil, err
	}

	// Update only the Hibernated field
	if updateIn.Hibernated != nil {
		module.Spec.Hibernated = *updateIn.Hibernated
	}

	return module, s.client.Update(ctx, module)
}

func (s ForkspacerModuleService) Delete(ctx context.Context, name string, namespace *string) error {
	if namespace == nil {
		namespace = utils.ToPtr("default")
	}

	module := &batchv1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: *namespace,
		},
	}

	return s.client.Delete(ctx, module)
}

func (s ForkspacerModuleService) List(
	ctx context.Context,
	limit int64,
	continueToken *string,
) (*batchv1.ModuleList, error) {
	options := []client.ListOption{
		client.Limit(limit),
	}

	if continueToken != nil {
		options = append(options, client.Continue(*continueToken))
	}

	modules := &batchv1.ModuleList{}
	err := s.client.List(ctx, modules, options...)

	return modules, err
}
