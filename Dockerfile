FROM node:20-alpine AS webbuild
WORKDIR /app
COPY cmd/shorturl/package.json cmd/shorturl/package-lock.json* ./
RUN npm install
COPY cmd/shorturl/web/tailwind ./web/tailwind
COPY cmd/shorturl/web/templates ./web/templates
RUN npx tailwindcss -c ./web/tailwind/tailwind.config.js -i ./web/tailwind/input.css -o ./web/static/app.css --minify

FROM golang:1.23-bookworm AS gobuild
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=webbuild /app/web/static/app.css ./cmd/shorturl/web/static/app.css
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/shorturl ./cmd/shorturl

FROM debian:bookworm-slim
WORKDIR /
ENV HOST=0.0.0.0
ENV PORT=8080
ENV DB_PATH=/data/shorturl.db
VOLUME ["/data"]
COPY --from=gobuild /out/shorturl /shorturl
RUN mkdir -p /data
EXPOSE 8080
ENTRYPOINT ["/shorturl"]
