# syntax=docker/dockerfile:1
FROM oven/bun:1 AS base
WORKDIR /app

FROM base AS deps
COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile

# --- DEV STAGE ---
FROM base AS dev
COPY --from=deps /app/node_modules ./node_modules
COPY frontend/ .
EXPOSE 3000
CMD ["bun", "run", "dev", "--host", "0.0.0.0"]

# --- PRODUCTION STAGE ---
FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY frontend/ .
RUN bun run build

FROM base AS production
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY frontend/package.json frontend/vite.config.ts frontend/tsconfig.json ./
EXPOSE 3000
CMD ["bun", "run", "preview", "--host", "0.0.0.0"]
