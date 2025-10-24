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

type KubeconfigSecretResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
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
		response.JSONBadRequest(w, "Kubeconfig file is required")
		return
	}
	defer func() { _ = file.Close() }()

	requestData.Kubeconfig, err = io.ReadAll(file)
	if err != nil {
		response.JSONBadRequest(w, "Invalid kubeconfig file content: "+err.Error())
		return
	}

	if err := validation.FormDataBodyValidate(w, r, requestData); err != nil {
		return
	}

	if secret, err := h.forkspacerWorkspaceService.CreateKubeconfigSecret(
		r.Context(), requestData.Name, nil, requestData.Kubeconfig,
	); err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	} else {
		response.JSONSuccess(w, 201,
			response.NewJSONSuccess(
				response.SuccessCodes.Created,
				KubeconfigSecretResponse{
					Namespace: secret.Namespace,
					Name:      secret.Name,
				},
			),
		)
		return
	}
}

type DeleteKubeconfigSecretRequest struct {
	Name string `json:"name" validate:"required,dns1123subdomain"`
}

func (h WorkspaceHandler) DeleteKubeconfigSecretHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &DeleteKubeconfigSecretRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	if err := h.forkspacerWorkspaceService.DeleteKubeconfigSecret(
		r.Context(), requestData.Name, nil,
	); err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	} else {
		response.JSONDeleted(w)
		return
	}
}

type ListKubeconfigSecretsRequestQuery struct {
	Limit         *int64  `json:"limit" validate:"omitempty,gte=1,lte=250"`
	ContinueToken *string `json:"continueToken"`
}

type ListKubeconfigSecretsResponse struct {
	ContinueToken string                     `json:"continueToken"`
	Secrets       []KubeconfigSecretResponse `json:"secrets"`
}

func (h WorkspaceHandler) ListKubeconfigSecretsHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &ListKubeconfigSecretsRequestQuery{}

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

	if secrets, err := h.forkspacerWorkspaceService.ListKubeconfigSecrets(
		r.Context(), *requestData.Limit, requestData.ContinueToken,
	); err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	} else {
		responseData := ListKubeconfigSecretsResponse{
			ContinueToken: secrets.Continue,
			Secrets:       make([]KubeconfigSecretResponse, len(secrets.Items)),
		}
		for i, secret := range secrets.Items {
			responseData.Secrets[i] = KubeconfigSecretResponse{
				Namespace: secret.Namespace,
				Name:      secret.Name,
			}
		}

		response.JSONSuccess(w, 200,
			response.NewJSONSuccess(
				response.SuccessCodes.Ok,
				responseData,
			),
		)
		return
	}
}

type WorkspaceResourceReference struct {
	Name      string `json:"name" validate:"required,dns1123subdomain"`
	Namespace string `json:"namespace" validate:"required,dns1123label"`
}

type WorkspaceConnection struct {
	Type   string                      `json:"type" validate:"required,oneof=kubeconfig in-cluster"`
	Secret *WorkspaceResourceReference `json:"secret,omitempty" validate:"required_if=Type kubeconfig"`
	Key    *string                     `json:"key,omitempty" validate:"omitempty,min=1"`
}

type WorkspaceAutoHibernation struct {
	Enabled      bool    `json:"enabled"`
	Schedule     string  `json:"schedule" validate:"required_if=Enabled true,omitempty,cron"`
	WakeSchedule *string `json:"wakeSchedule,omitempty" validate:"omitempty,cron"`
}

type ManagedCluster struct {
	Backend *string `json:"backend,omitempty" validate:"omitempty,oneof=vcluster k3d kind"`
	Distro  *string `json:"distro,omitempty" validate:"omitempty,oneof=k3s k0s k8s eks"`
}

type CreateWorkspaceRequest struct {
	Name            string                      `json:"name" validate:"required,dns1123subdomain"`
	Namespace       *string                     `json:"namespace,omitempty" validate:"omitempty,dns1123label"`
	Type            *string                     `json:"type,omitempty" validate:"omitempty,oneof=kubernetes managed"`
	From            *WorkspaceResourceReference `json:"from,omitempty"`
	Hibernated      bool                        `json:"hibernated"`
	Connection      *WorkspaceConnection        `json:"connection" validate:"required"`
	ManagedCluster  *ManagedCluster             `json:"managedCluster,omitempty"`
	AutoHibernation *WorkspaceAutoHibernation   `json:"autoHibernation,omitempty"`
}

type WorkspaceResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (h WorkspaceHandler) CreateHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &CreateWorkspaceRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	workspaceIn := forkspacer.WorkspaceCreateIn{
		Name:       requestData.Name,
		Namespace:  requestData.Namespace,
		Type:       requestData.Type,
		Hibernated: requestData.Hibernated,
	}

	if requestData.From != nil {
		workspaceIn.From = &forkspacer.ResourceReference{
			Name:      requestData.From.Name,
			Namespace: requestData.From.Namespace,
		}
	}

	if requestData.Connection != nil {
		workspaceIn.Connection = &forkspacer.WorkspaceCreateConnectionIn{
			Type: requestData.Connection.Type,
			Key:  requestData.Connection.Key,
		}
		if requestData.Connection.Secret != nil {
			workspaceIn.Connection.Secret = &forkspacer.ResourceReference{
				Name:      requestData.Connection.Secret.Name,
				Namespace: requestData.Connection.Secret.Namespace,
			}
		}
	}

	if requestData.ManagedCluster != nil {
		workspaceIn.ManagedCluster = &forkspacer.ManagedClusterIn{
			Backend: requestData.ManagedCluster.Backend,
			Distro:  requestData.ManagedCluster.Distro,
		}
	}

	if requestData.AutoHibernation != nil {
		workspaceIn.AutoHibernation = &forkspacer.WorkspaceAutoHibernationIn{
			Enabled:      requestData.AutoHibernation.Enabled,
			Schedule:     requestData.AutoHibernation.Schedule,
			WakeSchedule: requestData.AutoHibernation.WakeSchedule,
		}
	}

	workspace, err := h.forkspacerWorkspaceService.Create(r.Context(), workspaceIn)
	if err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	response.JSONSuccess(w, 201,
		response.NewJSONSuccess(
			response.SuccessCodes.Created,
			WorkspaceResponse{
				Name:      workspace.Name,
				Namespace: workspace.Namespace,
			},
		),
	)
}

type UpdateWorkspaceRequest struct {
	Name            string                    `json:"name" validate:"required,dns1123subdomain"`
	Namespace       *string                   `json:"namespace,omitempty" validate:"omitempty,dns1123label"`
	Hibernated      *bool                     `json:"hibernated,omitempty"`
	AutoHibernation *WorkspaceAutoHibernation `json:"autoHibernation,omitempty"`
	ManagedCluster  *ManagedCluster           `json:"managedCluster,omitempty"`
}

func (h WorkspaceHandler) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &UpdateWorkspaceRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	updateIn := forkspacer.WorkspaceUpdateIn{
		Name:       requestData.Name,
		Namespace:  requestData.Namespace,
		Hibernated: requestData.Hibernated,
	}

	if requestData.AutoHibernation != nil {
		updateIn.AutoHibernation = &forkspacer.WorkspaceAutoHibernationIn{
			Enabled:      requestData.AutoHibernation.Enabled,
			Schedule:     requestData.AutoHibernation.Schedule,
			WakeSchedule: requestData.AutoHibernation.WakeSchedule,
		}
	}

	if requestData.ManagedCluster != nil {
		updateIn.ManagedCluster = &forkspacer.ManagedClusterIn{
			Backend: requestData.ManagedCluster.Backend,
			Distro:  requestData.ManagedCluster.Distro,
		}
	}

	workspace, err := h.forkspacerWorkspaceService.Update(r.Context(), updateIn)
	if err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	response.JSONSuccess(w, 200,
		response.NewJSONSuccess(
			response.SuccessCodes.Ok,
			WorkspaceResponse{
				Name:      workspace.Name,
				Namespace: workspace.Namespace,
			},
		),
	)
}

type DeleteWorkspaceRequest struct {
	Name      string  `json:"name" validate:"required,dns1123subdomain"`
	Namespace *string `json:"namespace,omitempty" validate:"omitempty,dns1123label"`
}

func (h WorkspaceHandler) DeleteHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &DeleteWorkspaceRequest{}
	if err := validation.JSONBodyReadAndValidate(w, r, requestData); err != nil {
		return
	}

	if err := h.forkspacerWorkspaceService.Delete(r.Context(), requestData.Name, requestData.Namespace); err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	response.JSONDeleted(w)
}

type ListWorkspacesRequestQuery struct {
	Limit         *int64  `json:"limit,omitempty" validate:"omitempty,gte=1,lte=250"`
	ContinueToken *string `json:"continueToken,omitempty"`
}

type WorkspaceListItem struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Phase      string `json:"phase"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Hibernated bool   `json:"hibernated"`
}

type ListWorkspacesResponse struct {
	ContinueToken string              `json:"continueToken"`
	Workspaces    []WorkspaceListItem `json:"workspaces"`
}

func (h WorkspaceHandler) ListHandle(w http.ResponseWriter, r *http.Request) {
	var requestData = &ListWorkspacesRequestQuery{}

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

	workspaceList, err := h.forkspacerWorkspaceService.List(
		r.Context(),
		*requestData.Limit,
		requestData.ContinueToken,
	)
	if err != nil {
		response.JSONBadRequest(w, err.Error())
		return
	}

	responseData := ListWorkspacesResponse{
		ContinueToken: workspaceList.Continue,
		Workspaces:    make([]WorkspaceListItem, len(workspaceList.Items)),
	}

	for i, workspace := range workspaceList.Items {
		hibernated := workspace.Spec.Hibernated

		if workspace.Status.Message == nil {
			workspace.Status.Message = utils.ToPtr("")
		}

		responseData.Workspaces[i] = WorkspaceListItem{
			Name:       workspace.Name,
			Namespace:  workspace.Namespace,
			Phase:      string(workspace.Status.Phase),
			Type:       string(workspace.Spec.Type),
			Hibernated: hibernated,
			Message:    *workspace.Status.Message,
		}
	}

	response.JSONSuccess(w, 200,
		response.NewJSONSuccess(
			response.SuccessCodes.Ok,
			responseData,
		),
	)
}
