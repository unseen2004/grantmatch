# GrantMatch AI

> AI-powered grant matching and application drafting for NGOs

## Problem
Small NGOs and nonprofits spend 100+ hours per year searching for grants and writing applications. Most lack the staff to do this effectively, leaving billions in available funding unclaimed.

## Solution
GrantMatch AI lets an NGO describe their mission in plain language. The system:
1. Embeds the description using Google Gemini `text-embedding-004`
2. Runs cosine similarity search against 100+ real grants from Grants.gov and other sources
3. Ranks results by semantic relevance with a match score
4. Generates a full, tailored application letter in one click using `gemini-1.5-flash`

## Live Demo
🔗 **[https://your-app.up.railway.app](https://your-app.up.railway.app)**

## Demo Video
▶️ **[Watch 3-minute walkthrough](https://youtube.com/your-link)**

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.23 + chi router |
| Vector DB | PostgreSQL 16 + pgvector |
| AI Embeddings | Google Gemini `text-embedding-004` (768 dims) |
| AI Generation | Google Gemini `gemini-1.5-flash` |
| Frontend | HTMX + Tailwind CSS (no React) |
| Deployment | Railway + Docker |

## Architecture

```
NGO Form Input
     ↓
Gemini Embedder → 768-dim vector
     ↓
pgvector cosine search (ivfflat index)
     ↓
Top 10 grant matches with scores
     ↓
Gemini Flash → tailored application draft
```

## Local Setup

### Prerequisites
- Docker & Docker Compose
- Google Gemini API key ([get one free](https://aistudio.google.com/))

### Run

```bash
git clone https://github.com/unseen2004/grantmatch
cd grantmatch
cp .env.example .env
# Add your GEMINI_API_KEY to .env
docker compose up --build
```

Open http://localhost:8080

## Social Impact
- Targets small NGOs (<10 staff) that cannot afford professional grant writers
- Reduces grant search time from days to seconds
- Democratizes access to public and institutional funding
