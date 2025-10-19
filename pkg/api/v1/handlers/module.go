package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/api/validation"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
	"github.com/forkspacer/api-server/pkg/utils"
	"go.uber.org/zap"
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

type ModuleSourceConfigMapRef struct {
	Name      string  `json:"name" validate:"required,dns1123subdomain"`
	Namespace string  `json:"namespace" validate:"required,dns1123label"`
	Key       *string `json:"key,omitempty" validate:"omitempty,min=1"`
}

type ModuleSourceChartRepository struct {
	URL     string  `json:"url" validate:"required,min=1"`
	Chart   string  `json:"chart" validate:"required,min=1"`
	Version *string `json:"version,omitempty"`
}

type ModuleSourceChartGit struct {
	Repo     string `json:"repo" validate:"required,min=1"`
	Path     string `json:"path" validate:"required,min=1"`
	Revision string `json:"revision" validate:"required,min=1"`
}

type ModuleSourceChartRef struct {
	ConfigMap  *ModuleSourceConfigMapRef    `json:"configMap,omitempty"`
	Repository *ModuleSourceChartRepository `json:"repository,omitempty"`
	Git        *ModuleSourceChartGit        `json:"git,omitempty"`
}

type ModuleSourceExistingHelmReleaseRef struct {
	Name        string               `json:"name" validate:"required,dns1123subdomain"`
	Namespace   string               `json:"namespace" validate:"required,dns1123label"`
	ChartSource ModuleSourceChartRef `json:"chartSource"`
	Values      map[string]any       `json:"values,omitempty"`
}

type ModuleSource struct {
	Raw                 *string                             `json:"raw,omitempty" validate:"omitempty,yaml"`
	HttpURL             *string                             `json:"httpURL,omitempty" validate:"omitempty,http_url"`
	ConfigMap           *ModuleSourceConfigMapRef           `json:"configMap,omitempty"`
	ExistingHelmRelease *ModuleSourceExistingHelmReleaseRef `json:"existingHelmRelease,omitempty"`
}

type CreateModuleRequest struct {
	Name       string             `json:"name" validate:"required,dns1123subdomain"`
	Namespace  *string            `json:"namespace,omitempty" validate:"omitempty,dns1123label"`
	Workspace  WorkspaceReference `json:"workspace"`
	Source     ModuleSource       `json:"source"`
	Config     map[string]any     `json:"config"`
	Hibernated bool               `json:"hibernated"`
}

type ModuleResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (h ModuleHandler) CreateHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &CreateModuleRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	// Validate that at least one source is provided
	sourceCount := 0
	if requestData.Source.Raw != nil {
		sourceCount++
	}
	if requestData.Source.HttpURL != nil {
		sourceCount++
	}
	if requestData.Source.ConfigMap != nil {
		sourceCount++
	}
	if requestData.Source.ExistingHelmRelease != nil {
		sourceCount++
	}

	if sourceCount == 0 {
		response.JSONBodyValidationError(w, map[string]string{
			"CreateModuleRequest.source": "At least one source must be provided: raw, httpURL, configMap, or existingHelmRelease.",
		})
		return
	}

	var sourceRawBytes []byte = nil
	if requestData.Source.Raw != nil {
		sourceRawBytes = []byte(*requestData.Source.Raw)
	}

	moduleSource := forkspacer.ModuleSource{
		Raw:     sourceRawBytes,
		HttpURL: requestData.Source.HttpURL,
	}

	if requestData.Source.ConfigMap != nil {
		moduleSource.ConfigMap = &forkspacer.ModuleSourceConfigMapRef{
			Name:      requestData.Source.ConfigMap.Name,
			Namespace: requestData.Source.ConfigMap.Namespace,
			Key:       requestData.Source.ConfigMap.Key,
		}
	}

	if requestData.Source.ExistingHelmRelease != nil {
		moduleSource.ExistingHelmRelease = &forkspacer.ModuleSourceExistingHelmReleaseRef{
			Name:      requestData.Source.ExistingHelmRelease.Name,
			Namespace: requestData.Source.ExistingHelmRelease.Namespace,
			Values:    requestData.Source.ExistingHelmRelease.Values,
		}

		// Set chart source
		if requestData.Source.ExistingHelmRelease.ChartSource.ConfigMap != nil {
			moduleSource.ExistingHelmRelease.ChartSource.ConfigMap = &forkspacer.ModuleSourceConfigMapRef{
				Name:      requestData.Source.ExistingHelmRelease.ChartSource.ConfigMap.Name,
				Namespace: requestData.Source.ExistingHelmRelease.ChartSource.ConfigMap.Namespace,
				Key:       requestData.Source.ExistingHelmRelease.ChartSource.ConfigMap.Key,
			}
		}

		if requestData.Source.ExistingHelmRelease.ChartSource.Repository != nil {
			moduleSource.ExistingHelmRelease.ChartSource.Repository = &forkspacer.ModuleSourceChartRepository{
				URL:     requestData.Source.ExistingHelmRelease.ChartSource.Repository.URL,
				Chart:   requestData.Source.ExistingHelmRelease.ChartSource.Repository.Chart,
				Version: requestData.Source.ExistingHelmRelease.ChartSource.Repository.Version,
			}
		}

		if requestData.Source.ExistingHelmRelease.ChartSource.Git != nil {
			moduleSource.ExistingHelmRelease.ChartSource.Git = &forkspacer.ModuleSourceChartGit{
				Repo:     requestData.Source.ExistingHelmRelease.ChartSource.Git.Repo,
				Path:     requestData.Source.ExistingHelmRelease.ChartSource.Git.Path,
				Revision: requestData.Source.ExistingHelmRelease.ChartSource.Git.Revision,
			}
		}
	}

	module, err := h.forkspacerModuleService.Create(r.Context(), forkspacer.ModuleCreateIn{
		Name:      requestData.Name,
		Namespace: requestData.Namespace,
		Workspace: forkspacer.ResourceReference{
			Name:      requestData.Workspace.Name,
			Namespace: requestData.Workspace.Namespace,
		},
		Source:     moduleSource,
		Config:     requestData.Config,
		Hibernated: requestData.Hibernated,
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

		// Extract module type from source
		moduleType := "Unknown"
		if module.Spec.Source.Raw != nil && len(module.Spec.Source.Raw.Raw) > 0 {
			var sourceData map[string]any
			if err := json.Unmarshal(module.Spec.Source.Raw.Raw, &sourceData); err == nil {
				if kind, ok := sourceData["kind"].(string); ok {
					moduleType = kind
				}
			}
		} else if module.Spec.Source.HttpURL != nil {
			moduleType = "Remote"
		}

		responseData.Modules[i] = ModuleListItem{
			Name:       module.Name,
			Namespace:  module.Namespace,
			Phase:      string(module.Status.Phase),
			Hibernated: hibernated,
			Message:    message,
			Type:       moduleType,
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
