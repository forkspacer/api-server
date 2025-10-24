package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/api/validation"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
	"github.com/forkspacer/api-server/pkg/utils"
	batchv1 "github.com/forkspacer/forkspacer/api/v1"
	"go.uber.org/zap"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ModuleHandler struct {
	logger                  *zap.Logger
	forkspacerModuleService *forkspacer.ForkspacerModuleService
}

func NewModuleHandler(
	logger *zap.Logger,
	forkspacerModuleService *forkspacer.ForkspacerModuleService,
) *ModuleHandler {
	return &ModuleHandler{logger, forkspacerModuleService}
}

type WorkspaceReference struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
}

// Helm-related request types
type ModuleSpecHelmChartRepoAuth struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
}

type ModuleSpecHelmChartRepo struct {
	URL     string                       `json:"url" validate:"required,min=1"`
	Chart   string                       `json:"chart" validate:"required,min=1"`
	Version *string                      `json:"version,omitempty"`
	Auth    *ModuleSpecHelmChartRepoAuth `json:"auth,omitempty"`
}

type ModuleSpecHelmChartConfigMap struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
	Key       string `json:"key" validate:"required,min=1"`
}

type ModuleSpecHelmChartGitAuthSecret struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
}

type ModuleSpecHelmChartGitAuth struct {
	HTTPSSecretRef *ModuleSpecHelmChartGitAuthSecret `json:"httpsSecretRef,omitempty"`
}

type ModuleSpecHelmChartGit struct {
	Repo     string                      `json:"repo" validate:"required,min=1"`
	Path     string                      `json:"path" validate:"required,min=1"`
	Revision string                      `json:"revision" validate:"required,min=1"`
	Auth     *ModuleSpecHelmChartGitAuth `json:"auth,omitempty"`
}

type ModuleSpecHelmChart struct {
	Repo      *ModuleSpecHelmChartRepo      `json:"repo,omitempty"`
	ConfigMap *ModuleSpecHelmChartConfigMap `json:"configMap,omitempty"`
	Git       *ModuleSpecHelmChartGit       `json:"git,omitempty"`
}

type ModuleSpecHelmExistingRelease struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
}

type ModuleSpecHelmValuesConfigMap struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
	Key       string `json:"key" validate:"required,min=1"`
}

type ModuleSpecHelmValues struct {
	File      *string                        `json:"file,omitempty"`
	ConfigMap *ModuleSpecHelmValuesConfigMap `json:"configMap,omitempty"`
	Raw       map[string]any                 `json:"raw,omitempty"`
}

type ModuleSpecHelmOutputValueFromSecret struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
	Key       string `json:"key" validate:"required,min=1"`
}

type ModuleSpecHelmOutputValueFrom struct {
	Secret *ModuleSpecHelmOutputValueFromSecret `json:"secret,omitempty"`
}

type ModuleSpecHelmOutput struct {
	Name      string                         `json:"name" validate:"required,min=1"`
	Value     any                            `json:"value,omitempty"`
	ValueFrom *ModuleSpecHelmOutputValueFrom `json:"valueFrom,omitempty"`
}

type ModuleSpecHelmCleanup struct {
	RemoveNamespace bool `json:"removeNamespace"`
	RemovePVCs      bool `json:"removePVCs"`
}

type ModuleSpecHelmMigration struct {
	PVCs       []string `json:"pvcs,omitempty"`
	ConfigMaps []string `json:"configMaps,omitempty"`
	Secrets    []string `json:"secrets,omitempty"`
}

type ModuleSpecHelm struct {
	ExistingRelease *ModuleSpecHelmExistingRelease `json:"existingRelease,omitempty"`
	Chart           ModuleSpecHelmChart            `json:"chart"`
	Namespace       string                         `json:"namespace" validate:"required,dns1123label"`
	Values          []ModuleSpecHelmValues         `json:"values,omitempty"`
	Outputs         []ModuleSpecHelmOutput         `json:"outputs,omitempty"`
	Cleanup         ModuleSpecHelmCleanup          `json:"cleanup"`
	Migration       ModuleSpecHelmMigration        `json:"migration"`
}

// Custom module request types
type ModuleSpecCustom struct {
	Image            string   `json:"image" validate:"required,min=1"`
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	Permissions      []string `json:"permissions,omitempty"`
}

// Config schema request types
type ConfigItemSpecInteger struct {
	Required bool `json:"required"`
	Default  int  `json:"default"`
	Min      *int `json:"min,omitempty"`
	Max      *int `json:"max,omitempty"`
	Editable bool `json:"editable"`
}

type ConfigItemSpecBoolean struct {
	Required bool `json:"required"`
	Default  bool `json:"default"`
	Editable bool `json:"editable"`
}

type ConfigItemSpecString struct {
	Required bool    `json:"required"`
	Default  string  `json:"default"`
	Regex    *string `json:"regex,omitempty"`
	Editable bool    `json:"editable"`
}

type ConfigItemSpecOption struct {
	Required bool     `json:"required"`
	Default  string   `json:"default"`
	Values   []string `json:"values" validate:"required,min=1"`
	Editable bool     `json:"editable"`
}

type ConfigItemSpecMultipleOptions struct {
	Required bool     `json:"required"`
	Default  []string `json:"default,omitempty"`
	Values   []string `json:"values" validate:"required,min=1"`
	Min      *int     `json:"min,omitempty"`
	Max      *int     `json:"max,omitempty"`
	Editable bool     `json:"editable"`
}

type ConfigItem struct {
	Name            string                         `json:"name" validate:"required,min=1"`
	Alias           string                         `json:"alias" validate:"required,min=1"`
	Integer         *ConfigItemSpecInteger         `json:"integer,omitempty"`
	Boolean         *ConfigItemSpecBoolean         `json:"boolean,omitempty"`
	String          *ConfigItemSpecString          `json:"string,omitempty"`
	Option          *ConfigItemSpecOption          `json:"option,omitempty"`
	MultipleOptions *ConfigItemSpecMultipleOptions `json:"multipleOptions,omitempty"`
}

type CreateModuleRequest struct {
	Name         string             `json:"name" validate:"required,dns1123subdomain"`
	Namespace    *string            `json:"namespace,omitempty" validate:"omitempty,dns1123label"`
	Workspace    WorkspaceReference `json:"workspace"`
	Helm         *ModuleSpecHelm    `json:"helm,omitempty"`
	Custom       *ModuleSpecCustom  `json:"custom,omitempty"`
	Config       map[string]any     `json:"config,omitempty"`
	ConfigSchema []ConfigItem       `json:"configSchema,omitempty"`
	Hibernated   bool               `json:"hibernated"`
}

type ModuleResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Conversion functions from handler types to CRD types
func convertHelmRequestToCRD(req *ModuleSpecHelm) *batchv1.ModuleSpecHelm {
	if req == nil {
		return nil
	}

	helm := &batchv1.ModuleSpecHelm{
		Namespace: req.Namespace,
		Chart:     convertHelmChartToCRD(req.Chart),
		Cleanup: batchv1.ModuleSpecHelmCleanup{
			RemoveNamespace: req.Cleanup.RemoveNamespace,
			RemovePVCs:      req.Cleanup.RemovePVCs,
		},
		Migration: batchv1.ModuleSpecHelmMigration{
			PVCs:       req.Migration.PVCs,
			ConfigMaps: req.Migration.ConfigMaps,
			Secrets:    req.Migration.Secrets,
		},
	}

	if req.ExistingRelease != nil {
		helm.ExistingRelease = &batchv1.ModuleSpecHelmExistingRelease{
			Name:      req.ExistingRelease.Name,
			Namespace: req.ExistingRelease.Namespace,
		}
	}

	if req.Values != nil {
		helm.Values = make([]batchv1.ModuleSpecHelmValues, len(req.Values))
		for i, v := range req.Values {
			helm.Values[i] = convertHelmValuesToCRD(v)
		}
	}

	if req.Outputs != nil {
		helm.Outputs = make([]batchv1.ModuleSpecHelmOutput, len(req.Outputs))
		for i, o := range req.Outputs {
			helm.Outputs[i] = convertHelmOutputToCRD(o)
		}
	}

	return helm
}

func convertHelmChartToCRD(chart ModuleSpecHelmChart) batchv1.ModuleSpecHelmChart {
	crd := batchv1.ModuleSpecHelmChart{}

	if chart.Repo != nil {
		crd.Repo = &batchv1.ModuleSpecHelmChartRepo{
			URL:     chart.Repo.URL,
			Chart:   chart.Repo.Chart,
			Version: chart.Repo.Version,
		}
		if chart.Repo.Auth != nil {
			crd.Repo.Auth = &batchv1.ModuleSpecHelmChartRepoAuth{
				Name:      chart.Repo.Auth.Name,
				Namespace: chart.Repo.Auth.Namespace,
			}
		}
	}

	if chart.ConfigMap != nil {
		crd.ConfigMap = &batchv1.ModuleSpecHelmChartConfigMap{
			Name:      chart.ConfigMap.Name,
			Namespace: chart.ConfigMap.Namespace,
			Key:       chart.ConfigMap.Key,
		}
	}

	if chart.Git != nil {
		crd.Git = &batchv1.ModuleSpecHelmChartGit{
			Repo:     chart.Git.Repo,
			Path:     chart.Git.Path,
			Revision: chart.Git.Revision,
		}
		if chart.Git.Auth != nil && chart.Git.Auth.HTTPSSecretRef != nil {
			crd.Git.Auth = &batchv1.ModuleSpecHelmChartGitAuth{
				HTTPSSecretRef: &batchv1.ModuleSpecHelmChartGitAuthSecret{
					Name:      chart.Git.Auth.HTTPSSecretRef.Name,
					Namespace: chart.Git.Auth.HTTPSSecretRef.Namespace,
				},
			}
		}
	}

	return crd
}

func convertHelmValuesToCRD(values ModuleSpecHelmValues) batchv1.ModuleSpecHelmValues {
	crd := batchv1.ModuleSpecHelmValues{
		File: values.File,
	}

	if values.ConfigMap != nil {
		crd.ConfigMap = &batchv1.ModuleSpecHelmValuesConfigMap{
			Name:      values.ConfigMap.Name,
			Namespace: values.ConfigMap.Namespace,
			Key:       values.ConfigMap.Key,
		}
	}

	if values.Raw != nil {
		rawJSON, _ := json.Marshal(values.Raw)
		crd.Raw = &runtime.RawExtension{Raw: rawJSON}
	}

	return crd
}

func convertHelmOutputToCRD(output ModuleSpecHelmOutput) batchv1.ModuleSpecHelmOutput {
	crd := batchv1.ModuleSpecHelmOutput{
		Name: output.Name,
	}

	if output.Value != nil {
		valueJSON, _ := json.Marshal(output.Value)
		crd.Value = &apiextensionsv1.JSON{Raw: valueJSON}
	}

	if output.ValueFrom != nil && output.ValueFrom.Secret != nil {
		crd.ValueFrom = &batchv1.ModuleSpecHelmOutputValueFrom{
			Secret: &batchv1.ModuleSpecHelmOutputValueFromSecret{
				Name:      output.ValueFrom.Secret.Name,
				Namespace: output.ValueFrom.Secret.Namespace,
				Key:       output.ValueFrom.Secret.Key,
			},
		}
	}

	return crd
}

func convertCustomRequestToCRD(req *ModuleSpecCustom) *batchv1.ModuleSpecCustom {
	if req == nil {
		return nil
	}

	custom := &batchv1.ModuleSpecCustom{
		Image:            req.Image,
		ImagePullSecrets: req.ImagePullSecrets,
	}

	if req.Permissions != nil {
		custom.Permissions = make([]batchv1.CustomModulePermissionType, len(req.Permissions))
		for i, p := range req.Permissions {
			custom.Permissions[i] = batchv1.CustomModulePermissionType(p)
		}
	}

	return custom
}

func convertConfigSchemaRequestToCRD(items []ConfigItem) []batchv1.ConfigItem {
	if items == nil {
		return nil
	}

	crdItems := make([]batchv1.ConfigItem, len(items))
	for i, item := range items {
		crdItems[i] = batchv1.ConfigItem{
			Name:  item.Name,
			Alias: item.Alias,
		}

		if item.Integer != nil {
			crdItems[i].Integer = &batchv1.ConfigItemSpecInteger{
				Required: item.Integer.Required,
				Default:  item.Integer.Default,
				Min:      item.Integer.Min,
				Max:      item.Integer.Max,
				Editable: item.Integer.Editable,
			}
		}

		if item.Boolean != nil {
			crdItems[i].Boolean = &batchv1.ConfigItemSpecBoolean{
				Required: item.Boolean.Required,
				Default:  item.Boolean.Default,
				Editable: item.Boolean.Editable,
			}
		}

		if item.String != nil {
			crdItems[i].String = &batchv1.ConfigItemSpecString{
				Required: item.String.Required,
				Default:  item.String.Default,
				Regex:    item.String.Regex,
				Editable: item.String.Editable,
			}
		}

		if item.Option != nil {
			crdItems[i].Option = &batchv1.ConfigItemSpecOption{
				Required: item.Option.Required,
				Default:  item.Option.Default,
				Values:   item.Option.Values,
				Editable: item.Option.Editable,
			}
		}

		if item.MultipleOptions != nil {
			crdItems[i].MultipleOptions = &batchv1.ConfigItemSpecMultipleOptions{
				Required: item.MultipleOptions.Required,
				Default:  item.MultipleOptions.Default,
				Values:   item.MultipleOptions.Values,
				Min:      item.MultipleOptions.Min,
				Max:      item.MultipleOptions.Max,
				Editable: item.MultipleOptions.Editable,
			}
		}
	}

	return crdItems
}

func (h ModuleHandler) CreateHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &CreateModuleRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	// Validate that either Helm or Custom is provided, but not both
	if requestData.Helm == nil && requestData.Custom == nil {
		response.JSONBodyValidationError(w, map[string]string{
			"CreateModuleRequest": "Either 'helm' or 'custom' must be provided.",
		})
		return
	}

	if requestData.Helm != nil && requestData.Custom != nil {
		response.JSONBodyValidationError(w, map[string]string{
			"CreateModuleRequest": "Only one of 'helm' or 'custom' can be provided, not both.",
		})
		return
	}

	// Convert handler types to CRD types
	var helmSpec *batchv1.ModuleSpecHelm
	var customSpec *batchv1.ModuleSpecCustom
	var configSchema []batchv1.ConfigItem

	if requestData.Helm != nil {
		helmSpec = convertHelmRequestToCRD(requestData.Helm)
	}

	if requestData.Custom != nil {
		customSpec = convertCustomRequestToCRD(requestData.Custom)
	}

	if requestData.ConfigSchema != nil {
		configSchema = convertConfigSchemaRequestToCRD(requestData.ConfigSchema)
	}

	module, err := h.forkspacerModuleService.Create(r.Context(), forkspacer.ModuleCreateIn{
		Name:      requestData.Name,
		Namespace: requestData.Namespace,
		Workspace: forkspacer.ResourceReference{
			Name:      requestData.Workspace.Name,
			Namespace: requestData.Workspace.Namespace,
		},
		Helm:         helmSpec,
		Custom:       customSpec,
		Config:       requestData.Config,
		ConfigSchema: configSchema,
		Hibernated:   requestData.Hibernated,
	})
	if err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	response.JSONSuccess(w, 201,
		response.NewJSONSuccess(
			response.SuccessCodes.Created,
			ModuleResponse{
				Name:      module.Name,
				Namespace: module.Namespace,
			},
		),
	)
}

type UpdateModuleRequest struct {
	Name       string  `json:"name" validate:"required,dns1123subdomain"`
	Namespace  *string `json:"namespace,omitempty" validate:"omitempty,dns1123label"`
	Hibernated *bool   `json:"hibernated,omitempty"`
}

func (h ModuleHandler) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &UpdateModuleRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	updateIn := forkspacer.ModuleUpdateIn{
		Name:       requestData.Name,
		Namespace:  requestData.Namespace,
		Hibernated: requestData.Hibernated,
	}

	module, err := h.forkspacerModuleService.Update(r.Context(), updateIn)
	if err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	response.JSONSuccess(w, 200,
		response.NewJSONSuccess(
			response.SuccessCodes.Ok,
			ModuleResponse{
				Name:      module.Name,
				Namespace: module.Namespace,
			},
		),
	)
}

type DeleteModuleRequest struct {
	Name      string  `json:"name" validate:"required,dns1123subdomain"`
	Namespace *string `json:"namespace,omitempty" validate:"omitempty,dns1123label"`
}

func (h ModuleHandler) DeleteHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &DeleteModuleRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	if err := h.forkspacerModuleService.Delete(r.Context(), requestData.Name, requestData.Namespace); err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	response.JSONDeleted(w)
}

type ListModulesRequestQuery struct {
	Limit         *int64  `json:"limit,omitempty" validate:"omitempty,gte=1,lte=250"`
	ContinueToken *string `json:"continueToken,omitempty"`
}

type ModuleListItem struct {
	Name       string              `json:"name"`
	Namespace  string              `json:"namespace"`
	Phase      string              `json:"phase"`
	Message    string              `json:"message"`
	Hibernated bool                `json:"hibernated"`
	Type       string              `json:"type"`
	Workspace  *WorkspaceReference `json:"workspace,omitempty"`
}

type ListModulesResponse struct {
	ContinueToken string           `json:"continueToken"`
	Modules       []ModuleListItem `json:"modules"`
}

func (h ModuleHandler) ListHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &ListModulesRequestQuery{}

	if r.URL.Query().Has("limit") {
		qLimit, err := utils.ParseString[int64](r.URL.Query().Get("limit"))
		if err != nil {
			response.JSONBadRequest(w, err.Error())
			return
		}
		requestData.Limit = &qLimit
	}

	if r.URL.Query().Has("continueToken") {
		requestData.ContinueToken = utils.ToPtr(r.URL.Query().Get("continueToken"))
	}

	if err := validation.URLParamsValidate(r.Context(), w, requestData); err != nil {
		return
	}

	if requestData.Limit == nil {
		requestData.Limit = utils.ToPtr[int64](25)
	}

	moduleList, err := h.forkspacerModuleService.List(
		r.Context(),
		*requestData.Limit,
		requestData.ContinueToken,
	)
	if err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	responseData := ListModulesResponse{
		ContinueToken: moduleList.Continue,
		Modules:       make([]ModuleListItem, len(moduleList.Items)),
	}

	for i, module := range moduleList.Items {
		hibernated := module.Spec.Hibernated

		message := ""
		if module.Status.Message != nil {
			message = *module.Status.Message
		}

		responseData.Modules[i] = ModuleListItem{
			Name:       module.Name,
			Namespace:  module.Namespace,
			Phase:      string(module.Status.Phase),
			Hibernated: hibernated,
			Message:    message,
			Type:       module.Status.Source,
			Workspace: &WorkspaceReference{
				Name:      module.Spec.Workspace.Name,
				Namespace: module.Spec.Workspace.Namespace,
			},
		}
	}

	response.JSONSuccess(w, 200,
		response.NewJSONSuccess(
			response.SuccessCodes.Ok,
			responseData,
		),
	)
}
