# syntax=docker/dockerfile:1.7

ARG NODE_VERSION=22
ARG PLATFORM=linux/amd64

FROM --platform=${PLATFORM} node:${NODE_VERSION}-alpine AS builder
WORKDIR /app
COPY apps/frontend/package.json ./
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN pnpm install
COPY apps/frontend/. ./
RUN pnpm build

FROM --platform=${PLATFORM} nginx:1.27-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 4173
CMD ["nginx", "-g", "daemon off;"]
