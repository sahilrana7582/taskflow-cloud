package workspace

import (
	"net/http"

	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
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

	var req CreateRequest
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteError(w, r, err)
		return
	}

	ws, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, ws)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaces, err := h.service.List(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.ListResponse[WorkspaceResponse]{Items: workspaces})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	ws, err := h.service.Get(r.Context(), userID, workspaceID)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, ws)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	var req UpdateRequest
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteError(w, r, err)
		return
	}

	ws, err := h.service.Update(r.Context(), userID, workspaceID, req)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, ws)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	if err := h.service.Delete(r.Context(), userID, workspaceID); err != nil {
		utils.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	members, err := h.service.ListMembers(r.Context(), userID, workspaceID)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.ListResponse[MemberResponse]{Items: members})
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	var req AddMemberRequest
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteError(w, r, err)
		return
	}

	member, err := h.service.AddMember(r.Context(), userID, workspaceID, req)
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, member)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.RequireUserID(w, r)
	if !ok {
		return
	}

	workspaceID, err := utils.PathUUID(r, "workspaceID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	memberID, err := utils.PathUUID(r, "userID")
	if err != nil {
		utils.WriteError(w, r, err)
		return
	}

	if err := h.service.RemoveMember(r.Context(), userID, workspaceID, memberID); err != nil {
		utils.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
