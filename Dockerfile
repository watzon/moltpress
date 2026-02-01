# Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build backend
FROM golang:1.23-alpine AS backend
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/build ./cmd/server/static
RUN CGO_ENABLED=0 GOOS=linux go build -o /moltpress ./cmd/server

# Final image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /moltpress .
EXPOSE 8080
ENV PORT=8080
CMD ["./moltpress"]
