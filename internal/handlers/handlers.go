package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/unseen2004/grantmatch/internal/ai"
	"github.com/unseen2004/grantmatch/internal/models"
	"github.com/unseen2004/grantmatch/internal/rag"
	"github.com/unseen2004/grantmatch/internal/repository"
)

type Handlers struct {
	sessionRepo *repository.SessionRepository
	grantRepo   *repository.GrantRepository
	appRepo     *repository.ApplicationRepository
	embedder    *ai.Embedder
	generator   *ai.Generator
	matcher     *rag.Matcher
	templates   *template.Template
}

func NewHandlers(
	sessionRepo *repository.SessionRepository,
	grantRepo *repository.GrantRepository,
	appRepo *repository.ApplicationRepository,
	embedder *ai.Embedder,
	generator *ai.Generator,
	matcher *rag.Matcher,
	templates *template.Template,
) *Handlers {
	return &Handlers{
		sessionRepo: sessionRepo,
		grantRepo:   grantRepo,
		appRepo:     appRepo,
		embedder:    embedder,
		generator:   generator,
		matcher:     matcher,
		templates:   templates,
	}
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) ListGrants(w http.ResponseWriter, r *http.Request) {
	grants, err := h.grantRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to load grants: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.templates.ExecuteTemplate(w, "grants.html", map[string]any{"Grants": grants}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) HandleMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ngoName := r.FormValue("ngo_name")
	mission := r.FormValue("mission")
	region := r.FormValue("region")
	categories := strings.Split(r.FormValue("categories"), ",")
	budgetStr := r.FormValue("budget")

	if strings.TrimSpace(ngoName) == "" || strings.TrimSpace(mission) == "" {
		http.Error(w, "ngo_name and mission are required", http.StatusBadRequest)
		return
	}

	var budget *int64
	if b, err := strconv.ParseInt(budgetStr, 10, 64); err == nil {
		budget = &b
	}

	session, err := h.sessionRepo.Create(ctx, models.NgoSession{
		Name:       ngoName,
		Mission:    mission,
		Region:     region,
		Categories: categories,
		Budget:     budget,
	})
	if err != nil {
		http.Error(w, "failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	embText := ngoName + " " + mission + " " + strings.Join(categories, " ")
	vec, err := h.embedder.Embed(ctx, embText)
	if err != nil {
		http.Error(w, "embedding failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.sessionRepo.UpdateEmbedding(ctx, session.ID, vec); err != nil {
		http.Error(w, "failed to save embedding: "+err.Error(), http.StatusInternalServerError)
		return
	}

	matches, err := h.matcher.FindMatches(ctx, vec, 10)
	if err != nil {
		http.Error(w, "matching failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "results.html", map[string]any{
		"SessionID": session.ID,
		"Matches":   matches,
		"Empty":     len(matches) == 0,
	}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) HandleDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := r.FormValue("session_id")
	grantID := r.FormValue("grant_id")
	scoreStr := r.FormValue("score")

	if sessionID == "" || grantID == "" {
		http.Error(w, "session_id and grant_id are required", http.StatusBadRequest)
		return
	}

	grant, err := h.grantRepo.GetByID(ctx, grantID)
	if err != nil {
		http.Error(w, "grant not found: "+err.Error(), http.StatusNotFound)
		return
	}

	session, err := h.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		http.Error(w, "session not found: "+err.Error(), http.StatusNotFound)
		return
	}

	draft, err := h.generator.GenerateDraft(ctx, session.Mission, grant.Title, grant.Description, grant.Eligibility)
	if err != nil {
		http.Error(w, "draft generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var score float64
	if s, err := strconv.ParseFloat(scoreStr, 64); err == nil {
		score = s
	}

	_ = h.appRepo.Create(ctx, models.Application{
		SessionID: sessionID,
		GrantID:   grantID,
		Score:     score,
		DraftText: draft,
	})

	if err := h.templates.ExecuteTemplate(w, "draft.html", map[string]any{
		"Grant": grant,
		"Draft": draft,
	}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}
