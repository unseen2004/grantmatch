# GrantMatch AI — Technical Implementation Plan

> **Project:** AI Grant-Writing & Funding Matcher for NGOs  
> **Hackathon:** Quantum Sprint: For Social Good  
> **Target Prize:** 1st Place ($2,500) + Best AI Innovation Track ($250)  
> **Stack:** Go · Gemini API · pgvector · HTMX · Railway  

---

## Overview

GrantMatch AI is a web application that helps small NGOs and nonprofits:
1. Find open grants that match their mission using semantic vector search
2. Auto-generate a tailored grant application draft using an LLM
3. Score their eligibility per grant
4. Track deadlines with email reminders

The system is built as a Go monolith with a REST/HTMX frontend, a PostgreSQL + pgvector database, and the Google Gemini API for embeddings and text generation.

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      FRONTEND (HTMX)                    │
│  HTML served by Go · HTMX partial swaps · Tailwind CSS  │
└───────────────────────────┬─────────────────────────────┘
                            │ HTTP
┌───────────────────────────▼─────────────────────────────┐
│                     GO HTTP SERVER                      │
│  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  │
│  │  REST API    │  │  RAG Engine   │  │  Scraper     │  │
│  │  /api/v1     │  │  (search+gen) │  │  (grants DB) │  │
│  └──────────────┘  └───────────────┘  └──────────────┘  │
└──────────┬─────────────────┬──────────────────┬──────────┘
           │                 │                  │
    ┌──────▼──────┐  ┌───────▼──────┐  ┌────────▼───────┐
    │ PostgreSQL  │  │  Gemini API  │  │  External Grant │
    │ + pgvector  │  │  (Google AI) │  │  Sources (HTTP) │
    └─────────────┘  └──────────────┘  └────────────────┘
```

---

## Directory Structure

```
grantmatch/
├── cmd/
│   ├── server/
│   │   └── main.go              # HTTP server entry point
│   └── scraper/
│       └── main.go              # One-off scraper CLI
├── internal/
│   ├── config/
│   │   └── config.go            # Env var loading (godotenv)
│   ├── db/
│   │   ├── db.go                # pgx pool setup
│   │   └── migrations/
│   │       ├── 001_init.sql
│   │       ├── 002_grants.sql
│   │       └── 003_applications.sql
│   ├── models/
│   │   ├── grant.go
│   │   ├── ngo.go
│   │   └── application.go
│   ├── repository/
│   │   ├── grant_repo.go        # DB queries for grants
│   │   └── application_repo.go
│   ├── ai/
│   │   ├── embedder.go          # Gemini text-embedding-004
│   │   └── generator.go         # Gemini 1.5 Flash draft generation
│   ├── rag/
│   │   └── matcher.go           # Vector search + re-ranking
│   ├── scraper/
│   │   ├── scraper.go           # Orchestrator
│   │   ├── sources/
│   │   │   ├── ungrants.go      # UN grants scraper
│   │   │   ├── euportals.go     # EU funding portal scraper
│   │   │   └── usagov.go        # Grants.gov scraper
│   │   └── normalizer.go        # Normalize scraped data → Grant model
│   ├── handlers/
│   │   ├── home.go
│   │   ├── match.go             # POST /match — main flow
│   │   ├── draft.go             # POST /draft — generate application
│   │   └── grants.go            # GET /grants — browse all grants
│   └── templates/
│       ├── base.html
│       ├── home.html
│       ├── results.html         # HTMX partial: grant matches
│       ├── draft.html           # HTMX partial: generated draft
│       └── components/
│           ├── grant_card.html
│           └── score_badge.html
├── static/
│   ├── css/
│   │   └── style.css            # Tailwind output
│   └── js/
│       └── app.js               # Minimal JS (copy button, etc.)
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

---

## Database Schema

Run all migrations in order via `golang-migrate` or manually at startup.

### Migration 001 — Init

```sql
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Migration 002 — Grants

```sql
CREATE TABLE grants (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title         TEXT NOT NULL,
    description   TEXT NOT NULL,
    funder        TEXT NOT NULL,
    amount_min    BIGINT,
    amount_max    BIGINT,
    currency      TEXT DEFAULT 'USD',
    deadline      DATE,
    url           TEXT,
    region        TEXT,          -- e.g. "global", "europe", "sub-saharan-africa"
    categories    TEXT[],        -- e.g. ["health","education","climate"]
    eligibility   TEXT,          -- raw eligibility criteria text
    embedding     vector(768),   -- Gemini text-embedding-004 output (768 dims)
    source        TEXT,          -- which scraper populated this
    scraped_at    TIMESTAMPTZ DEFAULT now(),
    created_at    TIMESTAMPTZ DEFAULT now()
);

-- ANN index for fast similarity search
CREATE INDEX grants_embedding_idx ON grants
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX grants_deadline_idx ON grants (deadline);
CREATE INDEX grants_categories_idx ON grants USING GIN (categories);
```

### Migration 003 — Applications

```sql
CREATE TABLE ngo_sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          TEXT NOT NULL,
    mission       TEXT NOT NULL,
    region        TEXT,
    categories    TEXT[],
    budget        BIGINT,
    embedding     vector(768),
    created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE applications (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id    UUID REFERENCES ngo_sessions(id) ON DELETE CASCADE,
    grant_id      UUID REFERENCES grants(id) ON DELETE CASCADE,
    score         FLOAT NOT NULL,        -- 0.0 – 1.0 cosine similarity
    draft_text    TEXT,
    created_at    TIMESTAMPTZ DEFAULT now()
);
```

---

## Go Module & Dependencies

```
go mod init github.com/yourname/grantmatch
```

Required packages:

```
go get github.com/jackc/pgx/v5                  # PostgreSQL driver with pgvector support
go get github.com/pgvector/pgvector-go           # pgvector Go bindings
go get google.golang.org/genai                   # Gemini SDK (google-genai-go)
go get github.com/joho/godotenv                  # .env loading
go get github.com/go-chi/chi/v5                  # HTTP router
go get github.com/PuerkitoBio/goquery            # HTML scraping
go get github.com/golang-migrate/migrate/v4      # DB migrations
go get github.com/rs/zerolog                     # Structured logging
```

---

## Environment Variables

Create `.env` (copy from `.env.example`):

```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/grantmatch?sslmode=disable

# Google AI
GEMINI_API_KEY=your_gemini_api_key_here

# Server
PORT=8080
APP_ENV=development

# Scraper
SCRAPE_INTERVAL_HOURS=24
```

---

## Implementation Steps

### Step 1 — Project Bootstrap

1. Create `go.mod` with the module name.
2. Create `cmd/server/main.go`:
   - Load `.env` via `godotenv.Load()`
   - Initialize `config.Config` struct from env vars
   - Open pgx connection pool via `pgxpool.New()`
   - Run database migrations on startup
   - Mount chi router with handlers
   - Start HTTP server on `PORT`

3. Create `internal/config/config.go`:

```go
package config

import "os"

type Config struct {
    DatabaseURL      string
    GeminiAPIKey     string
    Port             string
    AppEnv           string
    ScrapeIntervalH  int
}

func Load() Config {
    return Config{
        DatabaseURL:  os.Getenv("DATABASE_URL"),
        GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
        Port:         getEnvOrDefault("PORT", "8080"),
        AppEnv:       getEnvOrDefault("APP_ENV", "development"),
    }
}

func getEnvOrDefault(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
```

---

### Step 2 — Database Layer

Create `internal/db/db.go`:

```go
package db

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connStr)
    if err != nil {
        return nil, err
    }
    config.MaxConns = 20
    return pgxpool.NewWithConfig(ctx, config)
}
```

Create `internal/db/migrations/` with all `.sql` files from the schema section above.

Create `internal/db/migrate.go` that runs all migration files in order at startup using `golang-migrate`.

---

### Step 3 — AI Embedder

Create `internal/ai/embedder.go`:

```go
package ai

import (
    "context"
    "google.golang.org/genai"
)

type Embedder struct {
    client *genai.Client
    model  string
}

func NewEmbedder(ctx context.Context, apiKey string) (*Embedder, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey: apiKey,
    })
    if err != nil {
        return nil, err
    }
    return &Embedder{client: client, model: "text-embedding-004"}, nil
}

// Embed generates a 768-dim vector for the given text.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
    result, err := e.client.Models.EmbedContent(ctx, e.model,
        genai.Text(text), nil)
    if err != nil {
        return nil, err
    }
    return result.Embeddings[0].Values, nil
}
```

**Important:** The embedding text for a grant should concatenate `title + " " + description + " " + eligibility`. The embedding text for an NGO should be `name + " " + mission + " " + categories`.

---

### Step 4 — AI Draft Generator

Create `internal/ai/generator.go`:

```go
package ai

import (
    "context"
    "fmt"
    "google.golang.org/genai"
)

type Generator struct {
    client *genai.Client
}

func NewGenerator(ctx context.Context, apiKey string) (*Generator, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
    if err != nil {
        return nil, err
    }
    return &Generator{client: client}, nil
}

func (g *Generator) GenerateDraft(ctx context.Context, ngoMission, grantTitle, grantDesc, eligibility string) (string, error) {
    prompt := fmt.Sprintf(`You are an expert grant writer. Write a compelling grant application letter for the following:

NGO MISSION: %s

GRANT: %s
GRANT DESCRIPTION: %s
ELIGIBILITY CRITERIA: %s

Write a professional 400-600 word application letter. Structure:
1. Opening paragraph: Who we are and why we are applying
2. Problem statement: The social problem we address
3. Proposed project: How we will use the grant
4. Impact metrics: Measurable outcomes we commit to
5. Closing: Why we are the ideal recipient

Be specific, persuasive, and align the NGO mission tightly with the grant criteria.`, ngoMission, grantTitle, grantDesc, eligibility)

    model := g.client.GenerativeModel("gemini-1.5-flash")
    resp, err := model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return "", err
    }
    return resp.Candidates[0].Content.Parts[0].(genai.Text).String(), nil
}
```

---

### Step 5 — RAG Matcher

Create `internal/rag/matcher.go`:

```go
package rag

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pgvector/pgvector-go"
)

type GrantMatch struct {
    GrantID     string
    Title       string
    Funder      string
    Description string
    Deadline    string
    AmountMax   int64
    URL         string
    Score       float64
}

type Matcher struct {
    pool *pgxpool.Pool
}

func NewMatcher(pool *pgxpool.Pool) *Matcher {
    return &Matcher{pool: pool}
}

// FindMatches performs cosine similarity search against the grants table.
// topK controls how many results to return.
func (m *Matcher) FindMatches(ctx context.Context, ngoEmbedding []float32, topK int) ([]GrantMatch, error) {
    query := `
        SELECT id, title, funder, description, 
               deadline, amount_max, url,
               1 - (embedding <=> $1) AS score
        FROM grants
        WHERE deadline >= CURRENT_DATE OR deadline IS NULL
        ORDER BY embedding <=> $1
        LIMIT $2
    `
    vec := pgvector.NewVector(ngoEmbedding)
    rows, err := m.pool.Query(ctx, query, vec, topK)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var matches []GrantMatch
    for rows.Next() {
        var m GrantMatch
        if err := rows.Scan(&m.GrantID, &m.Title, &m.Funder,
            &m.Description, &m.Deadline, &m.AmountMax, &m.URL, &m.Score); err != nil {
            continue
        }
        matches = append(matches, m)
    }
    return matches, nil
}
```

---

### Step 6 — Grant Scraper

Create `internal/scraper/scraper.go` as the orchestrator. Each source is a separate file in `internal/scraper/sources/`.

**Data sources to scrape (all publicly accessible):**

| Source | URL | Method |
|--------|-----|--------|
| Grants.gov (US) | `https://www.grants.gov/search-grants` | JSON API — use `https://apply07.grants.gov/grantsws/rest/opportunities/search/` with params |
| UN SDG Fund | `https://www.sdgfund.org/calls-for-proposals` | HTML scrape with goquery |
| EU Funding Portal | `https://ec.europa.eu/info/funding-tenders/opportunities/portal/screen/opportunities/topic-search` | JSON API endpoint |
| Candid (Foundation grants) | `https://candid.org/explore-issues/grants` | HTML scrape |

**Scraper interface (all sources implement this):**

```go
package scraper

import "context"

type RawGrant struct {
    Title       string
    Description string
    Funder      string
    AmountMin   int64
    AmountMax   int64
    Deadline    string   // YYYY-MM-DD or empty
    URL         string
    Region      string
    Categories  []string
    Eligibility string
    Source      string
}

type Source interface {
    Name() string
    Scrape(ctx context.Context) ([]RawGrant, error)
}
```

**Scraper orchestrator logic (in `scraper.go`):**

```go
func Run(ctx context.Context, sources []Source, embedder *ai.Embedder, repo *repository.GrantRepository) error {
    for _, source := range sources {
        grants, err := source.Scrape(ctx)
        if err != nil {
            log.Error().Err(err).Str("source", source.Name()).Msg("scrape failed")
            continue
        }
        for _, g := range grants {
            // 1. Normalize the grant
            normalized := Normalize(g)
            // 2. Generate embedding
            embeddingText := normalized.Title + " " + normalized.Description + " " + normalized.Eligibility
            vec, err := embedder.Embed(ctx, embeddingText)
            if err != nil {
                continue
            }
            normalized.Embedding = vec
            // 3. Upsert into DB (skip if URL already exists)
            _ = repo.Upsert(ctx, normalized)
        }
    }
    return nil
}
```

**Run the scraper as a background goroutine in `main.go`** on a ticker set to `SCRAPE_INTERVAL_HOURS`. Also expose `GET /admin/scrape` to trigger it manually.

---

### Step 7 — HTTP Handlers

Create `internal/handlers/match.go`:

```go
// POST /match
// Form fields: ngo_name, mission, region, categories (comma-sep), budget
func (h *MatchHandler) HandleMatch(w http.ResponseWriter, r *http.Request) {
    // 1. Parse form
    ngoName    := r.FormValue("ngo_name")
    mission    := r.FormValue("mission")
    region     := r.FormValue("region")
    categories := strings.Split(r.FormValue("categories"), ",")
    
    // 2. Save NGO session to DB
    session, _ := h.sessionRepo.Create(ctx, models.NgoSession{...})
    
    // 3. Embed NGO description
    embText := ngoName + " " + mission + " " + strings.Join(categories, " ")
    vec, _ := h.embedder.Embed(ctx, embText)
    
    // 4. Save embedding to session
    h.sessionRepo.UpdateEmbedding(ctx, session.ID, vec)
    
    // 5. Find top 10 matching grants
    matches, _ := h.matcher.FindMatches(ctx, vec, 10)
    
    // 6. Render HTMX partial (results.html)
    h.templates.ExecuteTemplate(w, "results.html", map[string]any{
        "SessionID": session.ID,
        "Matches":   matches,
    })
}
```

Create `internal/handlers/draft.go`:

```go
// POST /draft
// Form fields: session_id, grant_id
func (h *DraftHandler) HandleDraft(w http.ResponseWriter, r *http.Request) {
    sessionID := r.FormValue("session_id")
    grantID   := r.FormValue("grant_id")
    
    // 1. Fetch grant and session from DB
    grant,   _ := h.grantRepo.GetByID(ctx, grantID)
    session, _ := h.sessionRepo.GetByID(ctx, sessionID)
    
    // 2. Generate draft
    draft, _ := h.generator.GenerateDraft(ctx,
        session.Mission, grant.Title, grant.Description, grant.Eligibility)
    
    // 3. Save draft to applications table
    h.appRepo.Create(ctx, models.Application{
        SessionID: sessionID,
        GrantID:   grantID,
        Score:     0,     // already computed at match step
        DraftText: draft,
    })
    
    // 4. Render HTMX partial (draft.html)
    h.templates.ExecuteTemplate(w, "draft.html", map[string]any{
        "Grant": grant,
        "Draft": draft,
    })
}
```

---

### Step 8 — Router Setup

In `cmd/server/main.go`, mount all routes with chi:

```go
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// Static files
r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

// Pages
r.Get("/",           handlers.Home(tmpl))
r.Get("/grants",     handlers.ListGrants(grantRepo, tmpl))

// HTMX endpoints
r.Post("/match",     matchHandler.HandleMatch)
r.Post("/draft",     draftHandler.HandleDraft)

// Admin
r.Get("/admin/scrape", handlers.TriggerScrape(scraperOrchestrator))

http.ListenAndServe(":"+cfg.Port, r)
```

---

### Step 9 — Frontend (HTMX + Tailwind)

**Base layout (`internal/templates/base.html`)** — include:
- Tailwind CSS via CDN (`https://cdn.tailwindcss.com`)
- HTMX via CDN (`https://unpkg.com/htmx.org@1.9.12`)
- Dark mode toggle

**Home page form (`internal/templates/home.html`)**:

```html
<form hx-post="/match"
      hx-target="#results"
      hx-swap="innerHTML"
      hx-indicator="#spinner">
  
  <input name="ngo_name"   placeholder="Your NGO name" required />
  <textarea name="mission" placeholder="Describe your mission (2-3 sentences)" required></textarea>
  
  <select name="region">
    <option value="global">Global</option>
    <option value="europe">Europe</option>
    <option value="africa">Africa</option>
    <option value="asia">Asia</option>
    <option value="americas">Americas</option>
  </select>
  
  <input name="categories" placeholder="e.g. health, education, climate" />
  <input name="budget" type="number" placeholder="Budget needed (USD)" />
  
  <button type="submit">Find Matching Grants</button>
</form>

<div id="spinner" class="htmx-indicator">Searching grants…</div>
<div id="results"></div>
```

**Results partial (`internal/templates/results.html`)**:

```html
{{ range .Matches }}
<div class="grant-card">
  <div class="score-badge">{{ printf "%.0f" (mul .Score 100) }}% match</div>
  <h3>{{ .Title }}</h3>
  <p class="funder">{{ .Funder }}</p>
  <p>{{ truncate .Description 200 }}</p>
  <p>Deadline: {{ .Deadline }} | Up to: ${{ .AmountMax }}</p>
  
  <button hx-post="/draft"
          hx-vals='{"session_id": "{{ $.SessionID }}", "grant_id": "{{ .GrantID }}"}'
          hx-target="#draft-{{ .GrantID }}"
          hx-swap="innerHTML">
    Generate Application Draft
  </button>
  
  <div id="draft-{{ .GrantID }}"></div>
</div>
{{ end }}
```

**Draft partial (`internal/templates/draft.html`)**:

```html
<div class="draft-container">
  <h4>Draft Application for: {{ .Grant.Title }}</h4>
  <div class="draft-text" id="draft-text">{{ .Draft }}</div>
  <button onclick="copyDraft()">Copy to Clipboard</button>
  <a href="{{ .Grant.URL }}" target="_blank">View Original Grant →</a>
</div>
```

---

### Step 10 — Dockerfile

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o grantmatch ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/grantmatch .
COPY --from=builder /app/internal/templates ./internal/templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
EXPOSE 8080
CMD ["./grantmatch"]
```

---

### Step 11 — docker-compose.yml (Local Dev)

```yaml
version: '3.9'
services:
  db:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_USER: grantmatch
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: grantmatch
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  app:
    build: .
    ports:
      - "8080:8080"
    env_file: .env
    depends_on:
      - db

volumes:
  pgdata:
```

---

### Step 12 — Railway Deployment

1. Push code to GitHub.
2. Create a new Railway project → **Deploy from GitHub repo**.
3. Add a **PostgreSQL** plugin inside Railway — it auto-injects `DATABASE_URL`.
4. Enable the `pgvector` extension via the Railway Postgres console:
   ```sql
   CREATE EXTENSION IF NOT EXISTS vector;
   ```
5. Set environment variables in Railway dashboard:
   - `GEMINI_API_KEY`
   - `APP_ENV=production`
6. Railway auto-detects the `Dockerfile` and builds/deploys on every push.
7. Note the public Railway URL — add it to the Devpost submission.

---

## README Structure (for Devpost judges)

The README must include:

```markdown
# GrantMatch AI

> AI-powered grant matching and application drafting for NGOs

## Problem
Small NGOs spend 100+ hours/year searching for grants and writing applications...

## Solution
[Screenshot of the app]

## How It Works
1. NGO enters their mission and focus areas
2. System embeds the description and searches 500+ grants using cosine similarity
3. Top matches are ranked by semantic relevance score
4. One click generates a full, tailored application draft

## Tech Stack
- Go 1.23 (backend + HTTP server)
- PostgreSQL 16 + pgvector (vector similarity search)
- Google Gemini API (embeddings + text generation)
- HTMX (reactive frontend without React)
- Railway (deployment)

## Live Demo
https://your-app.up.railway.app

## Local Setup
...
```

---

## Build Order Checklist

Execute in this exact order:

- [ ] **Day 1 AM** — Bootstrap project, DB schema, migrations, docker-compose running locally
- [ ] **Day 1 PM** — Embedder + Generator wired to Gemini API, unit-tested with a hardcoded grant
- [ ] **Day 2 AM** — Scraper for Grants.gov JSON API (fastest source), populate 50+ grants in DB
- [ ] **Day 2 PM** — RAG matcher working end-to-end: embed NGO → query pgvector → return ranked grants
- [ ] **Day 3 AM** — All HTTP handlers + HTMX templates wired together, full flow working
- [ ] **Day 3 PM** — Deploy to Railway, verify live URL works, add 3 more grant sources
- [ ] **Day 4 AM** — Polish UI (Tailwind), write README, record 3-minute demo video
- [ ] **Day 4 PM** — Submit on Devpost with live URL, GitHub link, and demo video

---

## Demo Video Script (3 minutes)

1. **(0:00–0:20)** — Problem statement: "Small NGOs miss out on $X billion in grants annually because..."
2. **(0:20–0:50)** — Show the form: enter a real NGO mission (e.g. a climate NGO in Warsaw)
3. **(0:50–1:30)** — Show the ranked results appearing instantly with match scores
4. **(1:30–2:15)** — Click "Generate Draft" on the top match, show the full application letter appearing
5. **(2:15–2:45)** — Show the copy-to-clipboard, link to original grant, deadline visibility
6. **(2:45–3:00)** — "Deployed live at [URL] — source code at [GitHub]"

---

*End of technical plan.*
