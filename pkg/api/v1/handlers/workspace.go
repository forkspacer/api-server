package handlers

import (
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
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

func (h WorkspaceHandler) CreateKubeconfigSecretHandle(w http.ResponseWriter, r *http.Request) {
	if secret, err := h.forkspacerWorkspaceService.CreateKubeconfigSecret(
		r.Context(), "test", nil, []byte(``),
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
