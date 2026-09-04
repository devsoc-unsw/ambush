# Ambush

- `frontend/` — Vite + React SPA with TanStack Router (bun)
- `backend/` — Go HTTP API server

## Development
### Starting development
1. Start all services and prepare a dev container (dev container, Postgres):
```bash
docker compose up -d
```
If this doesn't work install the docker-compose plugin. To rebuild the image locally after Dockerfile changes, add `--build`.

2. To enter the dev environment
```bash
docker exec -it ambush-dev-container bash
```

### Closing down development
1. To exit the dev environment
```bash
exit
```

2. To close all services
```bash
docker compose down
```
In the dev shell: `cd frontend` then `bun run dev` (port 3000), `bun run lint` for lint, and `bun run test` for tests (test files need the suffix `.test.ts`).
For the backend: `cd backend` then `go run .` (see `backend/db/README.md` for the sqlc/goose/pgx database workflow).
