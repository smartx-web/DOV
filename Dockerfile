FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /dov ./cmd/server
FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /dov /app/dov
COPY --from=build /app/web /app/web
USER app
ENV GIN_MODE=release
EXPOSE 10000
CMD ["/app/dov"]
