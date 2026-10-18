package project

import (
	"net/http"

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

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	var req CreateRequest
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteError(w, r, err)
		return
	}

	p, err := h.service.Create(r.Context(), userID, workspaceID, req)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, p)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	projects, err := h.service.List(r.Context(), userID, workspaceID)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.ListResponse[models.Project]{Items: projects})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, projectID, err := pathIDs(r)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	p, err := h.service.Get(r.Context(), userID, workspaceID, projectID)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, projectID, err := pathIDs(r)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	var req UpdateRequest
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteError(w, r, err)
		return
	}

	p, err := h.service.Update(r.Context(), userID, workspaceID, projectID, req)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, projectID, err := pathIDs(r)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	if err := h.service.Delete(r.Context(), userID, workspaceID, projectID); err != nil {
		utils.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pathIDs(r *http.Request) (workspaceID, projectID uuid.UUID, err error) {
	workspaceID, err = utils.PathUUID(r, "workspaceID")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	projectID, err = utils.PathUUID(r, "projectID")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return workspaceID, projectID, nil
}
