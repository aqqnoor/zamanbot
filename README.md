
# Zamanbek — merged mono-repo

Structure:
- `backend/` (Go) — from back.zip
- `frontend/` (Vite + React) — from front.zip

## Quick start (dev)
1. Copy `.env` to repo root and set keys.
2. Terminal A:
```bash
cd backend
go mod tidy
go run .
```
3. Terminal B:
```bash
cd frontend
npm i
npm run dev
```

Open http://localhost:5173 and chat; requests go to `:8080` via `/api` proxy.

## Frontend config
`frontend/.env` contains:
```
VITE_API_BASE=/api
```

`vite.config.js` rewrites `/api` to backend target.

## Backend endpoints
- `POST /chat` → `{ reply, hints }`
- `POST /embed`, `POST /transcribe` (if used)
- `GET /products`, `GET /goals`, `GET /glossary`
- Swagger: `GET /swagger/index.html`

