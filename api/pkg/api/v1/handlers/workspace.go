package handlers

import (
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/services/forkspacer"
	"go.uber.org/zap"
)

type WorkspaceHandler struct {
	logger            *zap.Logger
	forkspacerService *forkspacer.ForkspacerService
}

func NewWorkspaceHandler(logger *zap.Logger, forkspacerService *forkspacer.ForkspacerService) *WorkspaceHandler {
	return &WorkspaceHandler{logger, forkspacerService}
}

func (h WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	response.JSONSuccess(w, 200, response.NewJSONSuccess(response.SuccessCodes.Ok, "Hello"))
}

func (h WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	response.JSONSuccess(w, 200, response.NewJSONSuccess(response.SuccessCodes.Ok, "Hello"))
}

func (h WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	response.JSONSuccess(w, 200, response.NewJSONSuccess(response.SuccessCodes.Ok, "Hello"))
}
