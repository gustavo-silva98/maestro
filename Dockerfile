# =============================================================================
# 🧱 Base Stage
# =============================================================================
FROM repo.intranet.pags/base-images/debian-slim:13 AS base
WORKDIR /app

RUN apt-get update && apt-get install -y golang
# Install trusted certificates
COPY --from=repo.intranet.pags/base-images/ca-certificates:latest /ca-certificates/ /usr/share/ca-certificates
COPY --from=repo.intranet.pags/base-images/ca-certificates:latest /ca-certificates.crt /etc/ssl/certs

# Set Go environment
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct

# =============================================================================
# 🛠 Dev Stage
# =============================================================================
FROM base AS dev

# Copy go.mod and go.sum separately to cache dependencies
COPY go.mod go.sum ./

# Installs dependencies
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download

# Copy source code
COPY . .

# Optional: install hot reload tool
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go install github.com/air-verse/air@latest

# Expose app port
EXPOSE 8080

# Default command for development (adjust e.g., native, air)
CMD ["go", "run", "cmd/api/main.go"]
#CMD ["air", "-c", ".air.toml"]

# =============================================================================
# 🧪 Test Stage
# =============================================================================
FROM dev AS test

# Run tests
CMD ["go", "test", "-v", "./..."]

# =============================================================================
# 🏗️ Build Stage
# =============================================================================
FROM base AS build

# Copy go.mod and go.sum separately to cache dependencies
COPY go.mod go.sum ./

# Installs dependencies
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download

# Copy source code
COPY . .

# Build static binary
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go build -ldflags="-s -w" -o /app/bin/server cmd/main.go

# =============================================================================
# 🚀 Release Stage
# =============================================================================
FROM repo.intranet.pags/gcr.io/distroless/static:nonroot AS release
WORKDIR /app

# Install trusted certificates
COPY --from=repo.intranet.pags/base-images/ca-certificates:latest /ca-certificates/ /usr/share/ca-certificates
COPY --from=repo.intranet.pags/base-images/ca-certificates:latest /ca-certificates.crt /etc/ssl/certs

# Copy final binary from build stage
COPY --from=build /app/bin/server /app/server

# Expose app port
EXPOSE 8080
# Default command
CMD ["/app/server"]