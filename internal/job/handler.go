package job

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, projectID, err := pathIDs(r, "projectID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	var req CreateRequest
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteError(w, r, err)
		return
	}

	j, err := h.service.Create(r.Context(), userID, workspaceID, projectID, req)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, j)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, projectID, err := pathIDs(r, "projectID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	query := r.URL.Query()
	params := ListParams{Cursor: query.Get("cursor")}

	if s := query.Get("status"); s != "" {
		status := models.JobStatus(s)
		params.Status = &status
	}

	if l := query.Get("limit"); l != "" {
		if params.Limit, err = strconv.Atoi(l); err != nil {
			utils.WriteError(w, r, utils.Invalid("limit must be a number"))
			return
		}
	}

	page, err := h.service.List(r.Context(), userID, workspaceID, projectID, params)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, page)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, jobID, err := pathIDs(r, "jobID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	j, err := h.service.Get(r.Context(), userID, workspaceID, jobID)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, j)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, jobID, err := pathIDs(r, "jobID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	var req UpdateRequest
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteError(w, r, err)
		return
	}

	j, err := h.service.Update(r.Context(), userID, workspaceID, jobID, req)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, j)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, jobID, err := pathIDs(r, "jobID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	if err := h.service.Delete(r.Context(), userID, workspaceID, jobID); err != nil {
		utils.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// pathIDs reads {workspaceID} plus one more UUID path parameter.
func pathIDs(r *http.Request, name string) (workspaceID, id uuid.UUID, err error) {
	workspaceID, err = utils.PathUUID(r, "workspaceID")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	id, err = utils.PathUUID(r, name)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return workspaceID, id, nil
}
