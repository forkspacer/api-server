package handlers

import (
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"go.uber.org/zap"
)

type WorkspaceHandler struct {
	logger *zap.Logger
}

func NewWorkspaceHandler(logger *zap.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{logger}
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
