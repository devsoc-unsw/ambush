# Ambush

- `frontend/` — Vite + React SPA with TanStack Router (bun)
- `backend/` — Go HTTP API server

## Development

If this is the first time you are starting development on this repo OR a change was made to a Dockerfile and you need to rebuild, first do:
```bash
docker compose up -d --build
```

However, if you have **already built** the Docker environment, you can forego the `--build`:
```bash
docker compose up -d
```
If this doesn't work install the docker-compose plugin.

To enter the dev environment:
```bash
docker exec -it ambush-dev-container bash
```

### Closing down development

1. To exit the dev environment:
```bash
exit
```

2. To close all services:
```bash
docker compose down
```

### Development

In the dev shell: `cd frontend` then `bun run dev` (port 3000), `bun run lint` for lint, and `bun run test` for tests (test files need the suffix `.test.ts`).

For the backend: `cd backend` then `go run .` (see `backend/db/README.md` for the sqlc/goose/pgx database workflow).

### Services and ports

| Service | Host port |
| --- | --- |
| frontend (vite) | 3000 |
| backend (go) | 8080 |
| metadata Postgres (`meta_ambush`) | 5432 |
| SpiceDB (gRPC) | 50067 |
| SpiceDB Postgres (`spice_ambush`) | 5467 |
