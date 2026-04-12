FROM node:24-bookworm-slim AS admin-builder

WORKDIR /src/web/admin

COPY web/admin/package.json web/admin/package-lock.json ./
RUN npm ci

COPY web/admin/ ./
RUN npm run build


FROM golang:1.24-bookworm AS go-builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY --from=admin-builder /src/web/admin/dist ./web/admin/dist

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/shorturl ./cmd/shorturl


FROM debian:bookworm-slim AS runtime

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /out/shorturl /app/shorturl
COPY --from=go-builder /src/web/admin/dist /app/web/admin/dist

RUN mkdir -p /data

ENV HOST=0.0.0.0
ENV PORT=8080
ENV DB_PATH=/data/shorturl.db
ENV ADMIN_STATIC_DIR=./web/admin/dist

EXPOSE 8080

CMD ["/app/shorturl"]
