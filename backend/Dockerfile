# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/scheduler ./cmd/scheduler

# Pre-create the media dir so that, when the shared apex-media-data volume first
# mounts over /media-data, Docker initializes the (empty) volume with this
# directory's nonroot ownership — letting the scheduler rehost images into it.
# The volume root is the catalog dir (see CATALOG_IMAGE_DIR); images go under
# cars/ and tracks/.
RUN mkdir -p /media-data/cars /media-data/tracks

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/scheduler /app/scheduler
COPY --from=build --chown=nonroot:nonroot /media-data /media-data
EXPOSE 8080
USER nonroot:nonroot
# Default: the API. The scheduler service overrides with `command: ["/app/scheduler"]`.
ENTRYPOINT ["/app/server"]
