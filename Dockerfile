# Build stage for Web Dashboard
FROM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build stage for Go CLI & Server
FROM golang:1.23-alpine AS go-builder
RUN apk add --no-cache build-base git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /codedna ./cmd/codedna

# Final runtime image
FROM alpine:3.20
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /codedna-app
COPY --from=go-builder /codedna /usr/local/bin/codedna
COPY --from=web-builder /app/web/out /codedna-app/web/out

EXPOSE 8080
ENTRYPOINT ["codedna"]
CMD ["serve", "--port", "8080", "--repo", "/workspace"]
