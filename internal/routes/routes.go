package routes

import (
	"net/http"

	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
	"github.com/sahilrana7582/taskflow-cloud/internal/job"
	"github.com/sahilrana7582/taskflow-cloud/internal/middleware"
	"github.com/sahilrana7582/taskflow-cloud/internal/project"
	"github.com/sahilrana7582/taskflow-cloud/internal/workspace"
)

// Handlers groups every HTTP handler the router needs.
type Handlers struct {
	Auth      *auth.AuthHandler
	Workspace *workspace.Handler
	Project   *project.Handler
	Job       *job.Handler
}

func NewServeMux(h Handlers, tokenManager *auth.TokenManager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", h.Auth.Register)
	mux.HandleFunc("/auth/login", h.Auth.Login)

	// Everything below is mounted under /api and requires a valid access token.
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /me", meHandler)

	protectedMux.HandleFunc("POST /workspaces", h.Workspace.Create)
	protectedMux.HandleFunc("GET /workspaces", h.Workspace.List)
	protectedMux.HandleFunc("GET /workspaces/{workspaceID}", h.Workspace.Get)
	protectedMux.HandleFunc("PATCH /workspaces/{workspaceID}", h.Workspace.Update)
	protectedMux.HandleFunc("DELETE /workspaces/{workspaceID}", h.Workspace.Delete)

	protectedMux.HandleFunc("GET /workspaces/{workspaceID}/members", h.Workspace.ListMembers)
	protectedMux.HandleFunc("POST /workspaces/{workspaceID}/members", h.Workspace.AddMember)
	protectedMux.HandleFunc("DELETE /workspaces/{workspaceID}/members/{userID}", h.Workspace.RemoveMember)

	protectedMux.HandleFunc("POST /workspaces/{workspaceID}/projects", h.Project.Create)
	protectedMux.HandleFunc("GET /workspaces/{workspaceID}/projects", h.Project.List)
	protectedMux.HandleFunc("GET /workspaces/{workspaceID}/projects/{projectID}", h.Project.Get)
	protectedMux.HandleFunc("PATCH /workspaces/{workspaceID}/projects/{projectID}", h.Project.Update)
	protectedMux.HandleFunc("DELETE /workspaces/{workspaceID}/projects/{projectID}", h.Project.Delete)

	// Jobs are created and listed per project, but addressed by workspace and
	// job ID alone, so a job URL does not depend on which project it is in.
	protectedMux.HandleFunc("POST /workspaces/{workspaceID}/projects/{projectID}/jobs", h.Job.Create)
	protectedMux.HandleFunc("GET /workspaces/{workspaceID}/projects/{projectID}/jobs", h.Job.List)
	protectedMux.HandleFunc("GET /workspaces/{workspaceID}/jobs/{jobID}", h.Job.Get)
	protectedMux.HandleFunc("PATCH /workspaces/{workspaceID}/jobs/{jobID}", h.Job.Update)
	protectedMux.HandleFunc("DELETE /workspaces/{workspaceID}/jobs/{jobID}", h.Job.Delete)

	mux.Handle(
		"/api/",
		http.StripPrefix(
			"/api",
			middleware.Auth(tokenManager)(protectedMux),
		),
	)

	handler := middleware.RequestID(mux)
	handler = middleware.Recovery(handler)

	return handler
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Authenticated user: " + userID.String()))
}
