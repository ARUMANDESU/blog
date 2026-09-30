FROM golang:1.27.1-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /blog ./cmd/blog

FROM scratch
WORKDIR /app
COPY --from=build /blog /app/blog
ENV SQLITE_DB_PATH=/app/data/db.sqlite
VOLUME [ "/app/data" ]
EXPOSE 8080
ENTRYPOINT ["/app/blog"]
