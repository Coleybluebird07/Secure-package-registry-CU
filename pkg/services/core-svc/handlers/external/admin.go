// Package external provides handlers for the external API.
package external

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"git.duti.dev/secure-package-registry/internal/gen/coredb"
	"git.duti.dev/secure-package-registry/internal/messages"
	"git.duti.dev/secure-package-registry/pkg/logger"
	sprminio "git.duti.dev/secure-package-registry/pkg/minio"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

// AdminHandler handles admin routes for package and task management.
type AdminHandler struct {
	db        coredb.Querier
	publisher message.Publisher
	minio     *sprminio.Client
	log       zerolog.Logger
}

// NewAdminHandler creates a new AdminHandler and registers its routes.
func NewAdminHandler(db coredb.Querier, publisher message.Publisher, minio *sprminio.Client) http.Handler {
	h := &AdminHandler{
		db:        db,
		publisher: publisher,
		minio:     minio,
		log:       logger.WithComponent("admin-handler"),
	}

	r := chi.NewRouter()
	r.Get("/packages", h.ListPackages)
	r.Post("/packages", h.AddPackage)
	r.Get("/packages/{ecosystem}/{identifier}/versions", h.ListVersions)
	r.Post("/packages/{ecosystem}/{identifier}/scan", h.TriggerScan)
	r.Get("/packages/{ecosystem}/{identifier}/behavior", h.GetBehavior)
	r.Get("/packages/{ecosystem}/{identifier}/behavior/raw", h.GetBehaviorRaw)
	r.Get("/packages/{ecosystem}/{identifier}/rebuild", h.GetRebuildForPackage)

	r.Get("/tasks", h.ListTasks)
	r.Get("/tasks/{taskID}/artifact", h.DownloadArtifact)

	r.Get("/rebuild-tasks", h.ListRebuildTasks)
	r.Get("/rebuild-tasks/{taskID}/artifact/{kind}", h.DownloadRebuildArtifact)

	return r
}

// ListPackages returns all watched packages for a given ecosystem.
func (h *AdminHandler) ListPackages(w http.ResponseWriter, r *http.Request) {
	ecoStr := r.URL.Query().Get("ecosystem")
	if ecoStr == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "missing required parameter: ecosystem"})
		return
	}

	ecosystem := coredb.Ecosystem(ecoStr)
	if !validEcosystem(ecosystem) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
		return
	}

	packages, err := h.db.ListPackagesByEcosystem(r.Context(), ecosystem)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list packages")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to list packages"})
		return
	}

	type packageResponse struct {
		ID            int32   `json:"id"`
		Identifier    string  `json:"identifier"`
		Ecosystem     string  `json:"ecosystem"`
		LatestVersion *string `json:"latest_version"`
	}

	items := make([]packageResponse, 0, len(packages))
	for _, pkg := range packages {
		resp := packageResponse{
			ID:         pkg.ID,
			Identifier: pkg.Identifier,
			Ecosystem:  string(pkg.Ecosystem),
		}
		if pkg.LatestVersion.Valid {
			resp.LatestVersion = &pkg.LatestVersion.String
		}
		items = append(items, resp)
	}

	render.JSON(w, r, map[string]any{"items": items})
}

// AddPackageRequest is the request body for AddPackage.
type AddPackageRequest struct {
	Identifier string `json:"identifier"`
	Ecosystem  string `json:"ecosystem"`
}

// AddPackage creates a new package in the watch list and publishes spr.package.requested.
func (h *AdminHandler) AddPackage(w http.ResponseWriter, r *http.Request) {
	var req AddPackageRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Identifier == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "missing required field: identifier"})
		return
	}
	if req.Ecosystem == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "missing required field: ecosystem"})
		return
	}

	ecosystem := coredb.Ecosystem(req.Ecosystem)
	if !validEcosystem(ecosystem) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + req.Ecosystem})
		return
	}

	ctx := r.Context()

	packageID, err := h.db.InsertPackage(ctx, coredb.InsertPackageParams{
		Identifier:    req.Identifier,
		Ecosystem:     ecosystem,
		LatestVersion: pgtype.Text{Valid: false},
	})

	alreadyExists := false
	if err != nil {
		h.log.Error().Err(err).Str("identifier", req.Identifier).Msg("Failed to insert package")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to create package"})
		return
	}

	if pubErr := h.publishPackageRequested(req.Identifier, req.Ecosystem); pubErr != nil {
		h.log.Error().Err(pubErr).Str("identifier", req.Identifier).Msg("Failed to publish package requested event")
	}

	status := http.StatusCreated
	if alreadyExists {
		status = http.StatusOK
	}

	render.Status(r, status)
	render.JSON(w, r, map[string]any{
		"id":             packageID,
		"identifier":     req.Identifier,
		"ecosystem":      req.Ecosystem,
		"already_exists": alreadyExists,
	})
}

// ListVersions returns all known versions for a package.
func (h *AdminHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	ecoStr := chi.URLParam(r, "ecosystem")
	identifier, _ := url.PathUnescape(chi.URLParam(r, "identifier"))

	ecosystem := coredb.Ecosystem(ecoStr)
	if !validEcosystem(ecosystem) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
		return
	}

	ctx := r.Context()

	pkg, err := h.db.GetPackageByEcosystemAndIdentifier(ctx, coredb.GetPackageByEcosystemAndIdentifierParams{
		Ecosystem:  ecosystem,
		Identifier: identifier,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "package not found"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).Str("identifier", identifier).Msg("Failed to look up package")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to look up package"})
		return
	}

	versions, err := h.db.ListPackageVersions(ctx, pkg.ID)
	if err != nil {
		h.log.Error().Err(err).Str("identifier", identifier).Msg("Failed to list versions")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to list versions"})
		return
	}

	var latestVersion *string
	if pkg.LatestVersion.Valid {
		latestVersion = &pkg.LatestVersion.String
	}

	render.JSON(w, r, map[string]any{
		"identifier":     identifier,
		"ecosystem":      ecoStr,
		"latest_version": latestVersion,
		"versions":       versions,
	})
}

// TriggerScanRequest is the optional request body for TriggerScan.
type TriggerScanRequest struct {
	Version string `json:"version"`
}

// TriggerScan triggers a behavioral analysis scan for a package version.
// If no version is specified, defaults to the stored latest_version.
func (h *AdminHandler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	ecoStr := chi.URLParam(r, "ecosystem")
	identifier, _ := url.PathUnescape(chi.URLParam(r, "identifier"))

	ecosystem := coredb.Ecosystem(ecoStr)
	if !validEcosystem(ecosystem) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
		return
	}

	ctx := r.Context()

	var req TriggerScanRequest
	_ = render.DecodeJSON(r.Body, &req)

	pkg, err := h.db.GetPackageByEcosystemAndIdentifier(ctx, coredb.GetPackageByEcosystemAndIdentifierParams{
		Ecosystem:  ecosystem,
		Identifier: identifier,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "package not found"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).Str("identifier", identifier).Msg("Failed to look up package")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to look up package"})
		return
	}

	version := req.Version
	if version == "" {
		if !pkg.LatestVersion.Valid {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "no version specified and package has no known latest version"})
			return
		}
		version = pkg.LatestVersion.String
	}

	pvID, err := h.db.InsertPackageVersion(ctx, coredb.InsertPackageVersionParams{
		PackageID: pkg.ID,
		Version:   version,
		SourceUrl: pgtype.Text{},
	})
	if err != nil {
		h.log.Error().Err(err).Str("identifier", identifier).Str("version", version).Msg("Failed to upsert package version")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to upsert package version"})
		return
	}

	task, err := h.db.InsertCollectionTask(ctx, coredb.InsertCollectionTaskParams{
		PackageVersionID: pvID,
		Source:           "npm",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, map[string]string{"error": "a collection task already exists for this version"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).Str("identifier", identifier).Str("version", version).Msg("Failed to insert collection task")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to create collection task"})
		return
	}

	collReq := messages.CollectionRequested{
		TaskID:     task.ID,
		Ecosystem:  ecoStr,
		Identifier: identifier,
		Version:    version,
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(collReq); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode collection request")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to encode collection request"})
		return
	}
	if err := h.publisher.Publish("spr.collection.requested", message.NewMessage(watermill.NewUUID(), buf.Bytes())); err != nil {
		h.log.Error().Err(err).Msg("Failed to publish collection request")
	}

	h.log.Info().
		Str("package", identifier).
		Str("version", version).
		Int32("task_id", task.ID).
		Msg("Triggered scan via admin API")

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]any{
		"task_id":    task.ID,
		"identifier": identifier,
		"ecosystem":  ecoStr,
		"version":    version,
		"status":     "pending",
	})
}

// GetBehavior returns the pre-computed deduped behavioral analysis tree
// for a specific package version.
func (h *AdminHandler) GetBehavior(w http.ResponseWriter, r *http.Request) {
	ecoStr := chi.URLParam(r, "ecosystem")
	identifier, _ := url.PathUnescape(chi.URLParam(r, "identifier"))

	ecosystem := coredb.Ecosystem(ecoStr)
	if !validEcosystem(ecosystem) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "missing required parameter: version"})
		return
	}

	ctx := r.Context()

	task, err := h.db.GetSucceededCollectionTask(ctx, coredb.GetSucceededCollectionTaskParams{
		Ecosystem:  ecosystem,
		Identifier: identifier,
		Version:    version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "no completed behavioral analysis found for this version"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).Str("identifier", identifier).Str("version", version).Msg("Failed to look up collection task")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to look up behavioral analysis"})
		return
	}

	rawKey := task.ArtifactKey.String
	dedupedKey := strings.TrimSuffix(rawKey, "behavior.jsonl") + "behavior-deduped.json"

	data, err := h.minio.GetObject(ctx, dedupedKey)
	if err != nil {
		h.log.Warn().Err(err).Str("key", dedupedKey).Msg("Deduped behavior tree not found in MinIO")
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "behavioral analysis data not yet processed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		h.log.Error().Err(err).Msg("Failed to write behavior response")
	}
}

// GetBehaviorRaw returns the pre-computed raw behavioral analysis tree.
func (h *AdminHandler) GetBehaviorRaw(w http.ResponseWriter, r *http.Request) {
	ecoStr := chi.URLParam(r, "ecosystem")
	identifier, _ := url.PathUnescape(chi.URLParam(r, "identifier"))

	ecosystem := coredb.Ecosystem(ecoStr)
	if !validEcosystem(ecosystem) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "missing required parameter: version"})
		return
	}

	ctx := r.Context()

	task, err := h.db.GetSucceededCollectionTask(ctx, coredb.GetSucceededCollectionTaskParams{
		Ecosystem:  ecosystem,
		Identifier: identifier,
		Version:    version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "no completed behavioral analysis found for this version"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).Str("identifier", identifier).Str("version", version).Msg("Failed to look up collection task")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to look up behavioral analysis"})
		return
	}

	rawKey := strings.TrimSuffix(task.ArtifactKey.String, "behavior.jsonl") + "behavior-raw.json"

	data, err := h.minio.GetObject(ctx, rawKey)
	if err != nil {
		h.log.Warn().Err(err).Str("key", rawKey).Msg("Raw behavior tree not found in MinIO")
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "raw behavioral analysis data not available"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		h.log.Error().Err(err).Msg("Failed to write raw behavior response")
	}
}

// ListTasks returns collection tasks with package context.
func (h *AdminHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	var ecosystem coredb.NullEcosystem
	if ecoStr := r.URL.Query().Get("ecosystem"); ecoStr != "" {
		eco := coredb.Ecosystem(ecoStr)
		if !validEcosystem(eco) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
			return
		}
		ecosystem = coredb.NullEcosystem{Ecosystem: eco, Valid: true}
	}

	page := int32(1)
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	pageSize := int32(50)
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = int32(ps)
		}
	}

	tasks, err := h.db.ListCollectionTasks(r.Context(), coredb.ListCollectionTasksParams{
		Ecosystem: ecosystem,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list collection tasks")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to list tasks"})
		return
	}

	type taskResponse struct {
		ID            int32   `json:"id"`
		Identifier    string  `json:"identifier"`
		Ecosystem     string  `json:"ecosystem"`
		Version       string  `json:"version"`
		Source        string  `json:"source"`
		Status        string  `json:"status"`
		FailureReason *string `json:"failure_reason,omitempty"`
		HasArtifact   bool    `json:"has_artifact"`
		StartedAt     *string `json:"started_at,omitempty"`
		CompletedAt   *string `json:"completed_at,omitempty"`
		CreatedAt     string  `json:"created_at"`
	}

	items := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp := taskResponse{
			ID:          t.ID,
			Identifier:  t.Identifier,
			Ecosystem:   t.PEcosystem,
			Version:     t.Version,
			Source:      t.Source,
			Status:      string(t.Status),
			HasArtifact: t.ArtifactBucket.Valid && t.ArtifactKey.Valid,
		}
		if t.FailureReason.Valid {
			resp.FailureReason = &t.FailureReason.String
		}
		if t.StartedAt.Valid {
			s := t.StartedAt.Time.String()
			resp.StartedAt = &s
		}
		if t.CompletedAt.Valid {
			s := t.CompletedAt.Time.String()
			resp.CompletedAt = &s
		}
		if t.CreatedAt.Valid {
			resp.CreatedAt = t.CreatedAt.Time.String()
		}
		items = append(items, resp)
	}

	render.JSON(w, r, map[string]any{"items": items})
}

// DownloadArtifact streams the behavioral analysis artifact for a completed collection task.
func (h *AdminHandler) DownloadArtifact(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid task ID"})
		return
	}

	task, err := h.db.GetCollectionTask(r.Context(), int32(taskID))
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "task not found"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).Int("task_id", taskID).Msg("Failed to get collection task")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to get task"})
		return
	}

	if !task.ArtifactBucket.Valid || !task.ArtifactKey.Valid {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "task has no artifact"})
		return
	}

	data, err := h.minio.GetObject(r.Context(), task.ArtifactKey.String)
	if err != nil {
		h.log.Error().Err(err).
			Str("bucket", task.ArtifactBucket.String).
			Str("key", task.ArtifactKey.String).
			Msg("Failed to download artifact from MinIO")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to download artifact"})
		return
	}

	filename := fmt.Sprintf("task-%d-behavior.jsonl", taskID)
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		h.log.Error().Err(err).Int("task_id", taskID).Msg("Failed to write artifact response")
	}
}

// ListRebuildTasks returns rebuild verification tasks with package context.
func (h *AdminHandler) ListRebuildTasks(w http.ResponseWriter, r *http.Request) {
	var ecosystem coredb.NullEcosystem
	if ecoStr := r.URL.Query().Get("ecosystem"); ecoStr != "" {
		eco := coredb.Ecosystem(ecoStr)
		if !validEcosystem(eco) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
			return
		}
		ecosystem = coredb.NullEcosystem{Ecosystem: eco, Valid: true}
	}

	page := int32(1)
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	pageSize := int32(50)
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = int32(ps)
		}
	}

	tasks, err := h.db.ListRebuildTasks(r.Context(), coredb.ListRebuildTasksParams{
		Ecosystem: ecosystem,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list rebuild tasks")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to list rebuild tasks"})
		return
	}

	type rebuildTaskResponse struct {
		ID            int32   `json:"id"`
		Identifier    string  `json:"identifier"`
		Ecosystem     string  `json:"ecosystem"`
		Version       string  `json:"version"`
		Source        string  `json:"source"`
		Status        string  `json:"status"`
		Matched       *bool   `json:"matched,omitempty"`
		FailureReason *string `json:"failure_reason,omitempty"`
		HasDiffoscope bool    `json:"has_diffoscope"`
		HasLogs       bool    `json:"has_logs"`
		HasMetadata   bool    `json:"has_metadata"`
		StartedAt     *string `json:"started_at,omitempty"`
		CompletedAt   *string `json:"completed_at,omitempty"`
		CreatedAt     string  `json:"created_at"`
	}

	items := make([]rebuildTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp := rebuildTaskResponse{
			ID:            t.ID,
			Identifier:    t.Identifier,
			Ecosystem:     t.PEcosystem,
			Version:       t.Version,
			Source:        t.Source,
			Status:        string(t.Status),
			HasDiffoscope: t.DiffoscopeBucket.Valid && t.DiffoscopeKey.Valid,
			HasLogs:       t.LogsBucket.Valid && t.LogsKey.Valid,
			HasMetadata:   t.MetadataBucket.Valid && t.MetadataKey.Valid,
		}
		if t.Matched.Valid {
			resp.Matched = &t.Matched.Bool
		}
		if t.FailureReason.Valid {
			resp.FailureReason = &t.FailureReason.String
		}
		if t.StartedAt.Valid {
			s := t.StartedAt.Time.String()
			resp.StartedAt = &s
		}
		if t.CompletedAt.Valid {
			s := t.CompletedAt.Time.String()
			resp.CompletedAt = &s
		}
		if t.CreatedAt.Valid {
			resp.CreatedAt = t.CreatedAt.Time.String()
		}
		items = append(items, resp)
	}

	render.JSON(w, r, map[string]any{"items": items})
}

// GetRebuildForPackage returns rebuild verification status for a package version.
func (h *AdminHandler) GetRebuildForPackage(w http.ResponseWriter, r *http.Request) {
	ecoStr := chi.URLParam(r, "ecosystem")
	identifier, _ := url.PathUnescape(chi.URLParam(r, "identifier"))

	ecosystem := coredb.Ecosystem(ecoStr)
	if !validEcosystem(ecosystem) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid ecosystem: " + ecoStr})
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "missing required parameter: version"})
		return
	}

	source := r.URL.Query().Get("source")
	if source == "" {
		source = "oss-rebuild"
	}

	task, err := h.db.GetRebuildTaskForPackageVersion(r.Context(), coredb.GetRebuildTaskForPackageVersionParams{
		Ecosystem:  ecosystem,
		Identifier: identifier,
		Version:    version,
		Source:     source,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "no rebuild verification found for this version"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).
			Str("identifier", identifier).
			Str("version", version).
			Str("source", source).
			Msg("Failed to look up rebuild task")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to look up rebuild verification"})
		return
	}

	render.JSON(w, r, rebuildTaskDetailResponse(task))
}

// DownloadRebuildArtifact streams a rebuild artifact/report from MinIO.
func (h *AdminHandler) DownloadRebuildArtifact(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid task ID"})
		return
	}

	kind := chi.URLParam(r, "kind")

	task, err := h.db.GetRebuildTask(r.Context(), int32(taskID))
	if errors.Is(err, pgx.ErrNoRows) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "rebuild task not found"})
		return
	}
	if err != nil {
		h.log.Error().Err(err).Int("task_id", taskID).Msg("Failed to get rebuild task")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to get rebuild task"})
		return
	}

	key, contentType, filename, ok := rebuildArtifactInfo(task, kind)
	if !ok {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "requested rebuild artifact is not available"})
		return
	}

	data, err := h.minio.GetObject(r.Context(), key)
	if err != nil {
		h.log.Error().Err(err).
			Int("task_id", taskID).
			Str("kind", kind).
			Str("key", key).
			Msg("Failed to download rebuild artifact from MinIO")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "failed to download rebuild artifact"})
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		h.log.Error().Err(err).Int("task_id", taskID).Msg("Failed to write rebuild artifact response")
	}
}

type rebuildTaskDetail struct {
	ID                  int32   `json:"id"`
	PackageVersionID    int32   `json:"package_version_id"`
	Source              string  `json:"source"`
	Status              string  `json:"status"`
	Matched             *bool   `json:"matched,omitempty"`
	FailureReason       *string `json:"failure_reason,omitempty"`
	HasOfficialArtifact bool    `json:"has_official_artifact"`
	HasRebuiltArtifact  bool    `json:"has_rebuilt_artifact"`
	HasDiffoscope       bool    `json:"has_diffoscope"`
	HasLogs             bool    `json:"has_logs"`
	HasMetadata         bool    `json:"has_metadata"`
	OfficialArtifactURL *string `json:"official_artifact_url,omitempty"`
	RebuiltArtifactURL  *string `json:"rebuilt_artifact_url,omitempty"`
	DiffoscopeURL       *string `json:"diffoscope_url,omitempty"`
	LogsURL             *string `json:"logs_url,omitempty"`
	MetadataURL         *string `json:"metadata_url,omitempty"`
	StartedAt           *string `json:"started_at,omitempty"`
	HeartbeatAt         *string `json:"heartbeat_at,omitempty"`
	CompletedAt         *string `json:"completed_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

func rebuildTaskDetailResponse(task coredb.RebuildTask) rebuildTaskDetail {
	resp := rebuildTaskDetail{
		ID:                  task.ID,
		PackageVersionID:    task.PackageVersionID,
		Source:              task.Source,
		Status:              string(task.Status),
		HasOfficialArtifact: task.OfficialArtifactBucket.Valid && task.OfficialArtifactKey.Valid,
		HasRebuiltArtifact:  task.RebuiltArtifactBucket.Valid && task.RebuiltArtifactKey.Valid,
		HasDiffoscope:       task.DiffoscopeBucket.Valid && task.DiffoscopeKey.Valid,
		HasLogs:             task.LogsBucket.Valid && task.LogsKey.Valid,
		HasMetadata:         task.MetadataBucket.Valid && task.MetadataKey.Valid,
	}

	if task.Matched.Valid {
		resp.Matched = &task.Matched.Bool
	}
	if task.FailureReason.Valid {
		resp.FailureReason = &task.FailureReason.String
	}
	if task.StartedAt.Valid {
		s := task.StartedAt.Time.String()
		resp.StartedAt = &s
	}
	if task.HeartbeatAt.Valid {
		s := task.HeartbeatAt.Time.String()
		resp.HeartbeatAt = &s
	}
	if task.CompletedAt.Valid {
		s := task.CompletedAt.Time.String()
		resp.CompletedAt = &s
	}
	if task.CreatedAt.Valid {
		resp.CreatedAt = task.CreatedAt.Time.String()
	}
	if task.UpdatedAt.Valid {
		resp.UpdatedAt = task.UpdatedAt.Time.String()
	}

	if resp.HasOfficialArtifact {
		u := fmt.Sprintf("/api/v1/admin/rebuild-tasks/%d/artifact/official", task.ID)
		resp.OfficialArtifactURL = &u
	}
	if resp.HasRebuiltArtifact {
		u := fmt.Sprintf("/api/v1/admin/rebuild-tasks/%d/artifact/rebuilt", task.ID)
		resp.RebuiltArtifactURL = &u
	}
	if resp.HasDiffoscope {
		u := fmt.Sprintf("/api/v1/admin/rebuild-tasks/%d/artifact/diffoscope", task.ID)
		resp.DiffoscopeURL = &u
	}
	if resp.HasLogs {
		u := fmt.Sprintf("/api/v1/admin/rebuild-tasks/%d/artifact/logs", task.ID)
		resp.LogsURL = &u
	}
	if resp.HasMetadata {
		u := fmt.Sprintf("/api/v1/admin/rebuild-tasks/%d/artifact/metadata", task.ID)
		resp.MetadataURL = &u
	}

	return resp
}

func rebuildArtifactInfo(task coredb.RebuildTask, kind string) (key string, contentType string, filename string, ok bool) {
	switch kind {
	case "official":
		if !task.OfficialArtifactKey.Valid {
			return "", "", "", false
		}
		return task.OfficialArtifactKey.String, "application/gzip", fmt.Sprintf("rebuild-task-%d-official.tgz", task.ID), true
	case "rebuilt":
		if !task.RebuiltArtifactKey.Valid {
			return "", "", "", false
		}
		return task.RebuiltArtifactKey.String, "application/gzip", fmt.Sprintf("rebuild-task-%d-rebuilt.tgz", task.ID), true
	case "diffoscope":
		if !task.DiffoscopeKey.Valid {
			return "", "", "", false
		}
		return task.DiffoscopeKey.String, "text/html", fmt.Sprintf("rebuild-task-%d-diffoscope.html", task.ID), true
	case "logs":
		if !task.LogsKey.Valid {
			return "", "", "", false
		}
		return task.LogsKey.String, "text/plain", fmt.Sprintf("rebuild-task-%d.log", task.ID), true
	case "metadata":
		if !task.MetadataKey.Valid {
			return "", "", "", false
		}
		return task.MetadataKey.String, "application/json", fmt.Sprintf("rebuild-task-%d-metadata.json", task.ID), true
	default:
		return "", "", "", false
	}
}

func (h *AdminHandler) publishPackageRequested(identifier, ecosystem string) error {
	req := messages.PackageRequest{
		Ecosystem:  ecosystem,
		Identifier: identifier,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(req); err != nil {
		return err
	}

	return h.publisher.Publish("spr.package.requested", message.NewMessage(watermill.NewUUID(), buf.Bytes()))
}

func validEcosystem(eco coredb.Ecosystem) bool {
	switch eco {
	case coredb.EcosystemNpm, coredb.EcosystemGo, coredb.EcosystemCargo, coredb.EcosystemPypi:
		return true
	default:
		return false
	}
}
