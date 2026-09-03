# Official images via AWS Public ECR (Docker Hub timeouts are common).
# ---------- Base ----------
FROM public.ecr.aws/docker/library/golang:1.25-alpine AS base
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# ---------- Dev (hot reload) ----------
FROM base AS dev
# Prebuilt binary install (not `go install`): recent air releases pull in a
# heavy transitive dependency tree (incl. gohugoio/hugo) and require a newer
# Go toolchain than this base image ships, which breaks a from-source build.
RUN apk add --no-cache curl wget && \
    curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b /usr/local/bin
COPY . .
EXPOSE 8080
# air watches source changes and auto-rebuilds+restarts, no image rebuild needed
CMD ["air", "-c", ".air.toml"]

# ---------- Build (compile binary for production) ----------
FROM base AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/api

# ---------- Prod (final, slim image) ----------
FROM public.ecr.aws/docker/library/alpine:3.19 AS prod
RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -H -u 10001 appuser
WORKDIR /app
COPY --from=build /app/server .
COPY --from=build /app/storage ./storage
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
ENTRYPOINT ["./server"]
