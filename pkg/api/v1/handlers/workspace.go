package handlers

import (
	"io"
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/api/validation"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
	"github.com/forkspacer/api-server/pkg/utils"
	"go.uber.org/zap"
)

type WorkspaceHandler struct {
	logger                     *zap.Logger
	forkspacerWorkspaceService *forkspacer.ForkspacerWorkspaceService
}

func NewWorkspaceHandler(
	logger *zap.Logger,
	forkspacerWorkspaceService *forkspacer.ForkspacerWorkspaceService,
) *WorkspaceHandler {
	return &WorkspaceHandler{logger, forkspacerWorkspaceService}
}

type CreateKubeconfigSecretRequest struct {
	Name       string `json:"name" validate:"required,dns1123subdomain"`
	Kubeconfig []byte `json:"kubeconfig" validate:"required,kubeconfig"`
}

func (h WorkspaceHandler) CreateKubeconfigSecretHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &CreateKubeconfigSecretRequest{}

	// upload of 10 MB.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.JSONFormDataTooLarge(w, utils.ToPtr[int64](10<<20))
		return
	}
	requestData.Name = r.PostFormValue("name")

	file, _, err := r.FormFile("kubeconfig")
	if err != nil {
		response.JSONError(w, 400, response.NewJSONError(response.ErrCodes.BodyValidation, "Kubeconfig file is required"))
		return
	}
	defer func() { _ = file.Close() }()

	requestData.Kubeconfig, err = io.ReadAll(file)
	if err != nil {
		response.JSONError(w, 400, response.NewJSONError(response.ErrCodes.BodyValidation, "Invalid kubeconfig file content"))
		return
	}

	if err := validation.FormDataBodyValidate(w, r, requestData); err != nil {
		return
	}

	if secret, err := h.forkspacerWorkspaceService.CreateKubeconfigSecret(
		r.Context(), requestData.Name, nil, requestData.Kubeconfig,
	); err != nil {
		response.JSONError(w, 400, response.NewJSONError(response.ErrCodes.BadRequest, err.Error()))
		return
	} else {
		response.JSONSuccess(w, 201, response.NewJSONSuccess(response.SuccessCodes.Created, secret.UID))
		return
	}
}

func (h WorkspaceHandler) CreateHandle(w http.ResponseWriter, r *http.Request) {
	response.JSONSuccess(w, 200, response.NewJSONSuccess(response.SuccessCodes.Ok, "Hello"))
}

func (h WorkspaceHandler) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	response.JSONSuccess(w, 200, response.NewJSONSuccess(response.SuccessCodes.Ok, "Hello"))
}

func (h WorkspaceHandler) DeleteHandle(w http.ResponseWriter, r *http.Request) {
	response.JSONSuccess(w, 200, response.NewJSONSuccess(response.SuccessCodes.Ok, "Hello"))
}
