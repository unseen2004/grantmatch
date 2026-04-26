# GrantMatch AI

> AI-powered grant matching and application drafting for NGOs

## Problem
Small NGOs spend 100+ hours/year searching for grants and writing applications, often missing out on funding due to lack of resources.

## Solution
GrantMatch AI is a web application that helps small NGOs and nonprofits:
1. Find open grants that match their mission using semantic vector search
2. Auto-generate a tailored grant application draft using an LLM
3. Score their eligibility per grant
4. Track deadlines with email reminders

## How It Works
1. NGO enters their mission and focus areas
2. System embeds the description and searches grants using cosine similarity
3. Top matches are ranked by semantic relevance score
4. One click generates a full, tailored application draft

## Tech Stack
- Go 1.23 (backend + HTTP server)
- PostgreSQL 16 + pgvector (vector similarity search)
- Google Gemini API (embeddings + text generation)
- HTMX (reactive frontend without React)
- TailwindCSS (styling)
- Railway (deployment)

## Local Setup
1. Clone the repo
2. Create `.env` from `.env.example` and set your `GEMINI_API_KEY`
3. Run `docker-compose up -d db` to start PostgreSQL
4. Run `go run ./cmd/server` to start the app
