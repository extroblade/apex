# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/scheduler ./cmd/scheduler

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/scheduler /app/scheduler
EXPOSE 8080
USER nonroot:nonroot
# Default: the API. The scheduler service overrides with `command: ["/app/scheduler"]`.
ENTRYPOINT ["/app/server"]
