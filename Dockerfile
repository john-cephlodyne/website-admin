# syntax=docker/dockerfile:1.4

# node builder image
ARG NODE_IMAGE
ARG GO_IMAGE

ARG DEPS=svelte5-only

FROM ${NODE_IMAGE} AS frontend-builder
ARG DEPS
WORKDIR /app
# Copy the workspace config, lockfile, and ALL package definitions
# from the base image's /deps directory into our current directory.
RUN cp -a /deps/. /app/
# Install dependencies using the pre-warmed cache.
RUN pnpm install --offline --frozen-lockfile
# Copy the application SOURCE CODE the local machine.
COPY ./frontend ./node-deps/${DEPS}
# Fail the build if any vulnerability (low, moderate, high, or critical) is found.
RUN pnpm audit --audit-level=low
# Run the build from inside the project directory for cleaner logs.
RUN cd /app/node-deps/${DEPS} && pnpm build

# go builder image
FROM ${GO_IMAGE} AS server-builder
ARG LOCAL_DEPLOY=false
ARG LOG_VALUES=false
WORKDIR /go-server
COPY cmd/server ./cmd/server
COPY internal/ ./internal/
COPY vendor/ ./vendor/
COPY go.mod .

RUN go install golang.org/x/vuln/cmd/govulncheck@latest
RUN govulncheck ./...

RUN if [ "$LOCAL_DEPLOY" = "true" ] ; then \
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s \
  -X 'website/internal/ev.cloudFlag=false' \
  -X 'website/internal/jot.sensitiveBuild=false' \
  -X 'website/internal/jot.logFormat=text'" \
  -o ./server ./cmd/server/main.go; \
  elif [ "$LOG_VALUES" = "true" ] ; then \
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s \
  -X 'website/internal/ev.cloudFlag=true' \
  -X 'website/internal/jot.sensitiveBuild=false' \
  -X 'website/internal/jot.logFormat=gcp'" \
  -o ./server ./cmd/server/main.go; \
  else \
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s \
  -X 'website/internal/ev.cloudFlag=true' \
  -X 'website/internal/jot.sensitiveBuild=true' \
  -X 'website/internal/jot.logFormat=gcp'" \
  -o ./server ./cmd/server/main.go; \
  fi

# final image
FROM scratch AS final
ARG DEPS
WORKDIR /app
# Copy CA certificates from the Go builder stage
COPY --from=server-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy built artifacts
COPY --from=frontend-builder /app/node-deps/${DEPS}/dist ./frontend/dist
COPY --from=server-builder /go-server/server ./server
# Run as an unprivileged, completely anonymous user instead of root.
USER 10001:10001

CMD [ "./server" ]
