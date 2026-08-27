# Ambush

- `frontend/` — Vite + React SPA with TanStack Router (bun)
- `backend/` — Go HTTP API server

## Development
### Starting development:
If it is the first time setting up the repo, build the docker environment:
```bash
docker compose -f .devcontainer/docker-compose.yml up -d --build
```
If this doesn't work install the docker-compose plugin.

If you have already set this up, you can use forego the `--build`:
```bash
docker compose -f .devcontainer/docker-compose.yml up -d --build
```

To enter the dev environment:
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
docker compose -f .devcontainer/docker-compose.yml down
```

### Development
In the dev shell: `cd frontend` then `bun run dev` (port 3000), `bun run lint` for lint, and `bun run test` for tests (test files need the suffix `.test.ts`).

For the backend: `cd backend` then `go run .` (see `backend/db/README.md` for the sqlc/goose/pgx database workflow).
