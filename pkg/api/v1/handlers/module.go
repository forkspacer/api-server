package handlers

import (
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/api/validation"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
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
	Raw     map[string]any `json:"raw" validate:"omitempty"`
	HttpURL *string        `json:"httpURL" validate:"omitempty,http_url"`
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

	module, err := h.forkspacerModuleService.Create(r.Context(), forkspacer.ModuleCreateIn{
		Name:      requestData.Name,
		Namespace: requestData.Namespace,
		Workspace: forkspacer.ResourceReference{
			Name:      requestData.Workspace.Name,
			Namespace: requestData.Workspace.Namespace,
		},
		Source: forkspacer.ModuleSource{
			Raw:     requestData.Source.Raw,
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

func (h ModuleHandler) UpdateHandle(w http.ResponseWriter, r *http.Request) {}
func (h ModuleHandler) DeleteHandle(w http.ResponseWriter, r *http.Request) {}
func (h ModuleHandler) ListHandle(w http.ResponseWriter, r *http.Request)   {}
