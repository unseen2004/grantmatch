package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/hackathon/grantmatch/internal/ai"
	"github.com/hackathon/grantmatch/internal/models"
	"github.com/hackathon/grantmatch/internal/rag"
	"github.com/hackathon/grantmatch/internal/repository"
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
	h.templates.ExecuteTemplate(w, "home.html", nil)
}

func (h *Handlers) ListGrants(w http.ResponseWriter, r *http.Request) {
	grants, _ := h.grantRepo.GetAll(r.Context())
	h.templates.ExecuteTemplate(w, "grants.html", map[string]any{
		"Grants": grants,
	})
}

func (h *Handlers) HandleMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ngoName := r.FormValue("ngo_name")
	mission := r.FormValue("mission")
	region := r.FormValue("region")
	categories := strings.Split(r.FormValue("categories"), ",")
	budgetStr := r.FormValue("budget")
	var budget *int64
	if b, err := strconv.ParseInt(budgetStr, 10, 64); err == nil {
		budget = &b
	}

	session, _ := h.sessionRepo.Create(ctx, models.NgoSession{
		Name:       ngoName,
		Mission:    mission,
		Region:     region,
		Categories: categories,
		Budget:     budget,
	})

	embText := ngoName + " " + mission + " " + strings.Join(categories, " ")
	vec, _ := h.embedder.Embed(ctx, embText)

	h.sessionRepo.UpdateEmbedding(ctx, session.ID, vec)

	matches, _ := h.matcher.FindMatches(ctx, vec, 10)

	h.templates.ExecuteTemplate(w, "results.html", map[string]any{
		"SessionID": session.ID,
		"Matches":   matches,
	})
}

func (h *Handlers) HandleDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := r.FormValue("session_id")
	grantID := r.FormValue("grant_id")

	grant, _ := h.grantRepo.GetByID(ctx, grantID)
	session, _ := h.sessionRepo.GetByID(ctx, sessionID)

	draft, _ := h.generator.GenerateDraft(ctx, session.Mission, grant.Title, grant.Description, grant.Eligibility)

	h.appRepo.Create(ctx, models.Application{
		SessionID: sessionID,
		GrantID:   grantID,
		Score:     0,
		DraftText: draft,
	})

	h.templates.ExecuteTemplate(w, "draft.html", map[string]any{
		"Grant": grant,
		"Draft": draft,
	})
}
