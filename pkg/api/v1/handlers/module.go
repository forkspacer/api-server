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

type ModuleSource struct {
	Raw     *string `json:"raw" validate:"omitempty,yaml"`
	HttpURL *string `json:"httpURL" validate:"omitempty,http_url"`
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

	if requestData.Source.HttpURL == nil && requestData.Source.Raw == nil {
		response.JSONBodyValidationError(w, map[string]string{
			"CreateModuleRequest.source": "At least one of 'raw' or 'httpURL' must be provided.",
		})
		return
	}

	var sourceRawBytes []byte = nil
	if requestData.Source.Raw != nil {
		sourceRawBytes = []byte(*requestData.Source.Raw)
	}

	module, err := h.forkspacerModuleService.Create(r.Context(), forkspacer.ModuleCreateIn{
		Name:      requestData.Name,
		Namespace: requestData.Namespace,
		Workspace: forkspacer.ResourceReference{
			Name:      requestData.Workspace.Name,
			Namespace: requestData.Workspace.Namespace,
		},
		Source: forkspacer.ModuleSource{
			Raw:     sourceRawBytes,
			HttpURL: requestData.Source.HttpURL,
		},
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
		hibernated := false
		if module.Spec.Hibernated != nil {
			hibernated = *module.Spec.Hibernated
		}

		message := ""
		if module.Status.Message != nil {
			message = *module.Status.Message
		}

		// Extract module type from source
		moduleType := "Unknown"
		if module.Spec.Source.Raw != nil && len(module.Spec.Source.Raw.Raw) > 0 {
			var sourceData map[string]interface{}
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
