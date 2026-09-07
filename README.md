# Ambush

- `frontend/` — Vite + React SPA with TanStack Router (bun)
- `backend/` — Go HTTP API server

## Development

### Starting development

1. Start all services — dev container, metadata Postgres, SpiceDB and its own Postgres:
```bash
docker compose up -d
```
If this doesn't work install the docker-compose plugin.

2. To enter the dev environment:
```bash
docker exec -it ambush-dev-container bash
```

### Rebuilding the dev container image

To rebuild the local image after `.devcontainer/Dockerfile` changes (retags `ghcr.io/devsoc-unsw/ambush-devcontainer` on your machine; does not push):
```bash
docker compose up -d --build
```

Build only:
```bash
docker compose build
```

To publish that image to GHCR, merge the Dockerfile change to `main`. CI then pushes multi-arch `ghcr.io/devsoc-unsw/ambush-devcontainer:latest`. From `main` you can also run:
```bash
gh workflow run devcontainer.yml --ref main
```

Do not `docker compose push` a local build — that would replace the multi-arch registry image with a single-platform one.

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
