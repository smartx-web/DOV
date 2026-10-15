FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /dov ./cmd/server && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /import-inec ./cmd/import-inec

# Third-party snapshot of the INEC locator, NOT an official INEC API.
# Pin to a fixed revision for reproducibility; update only after review.
FROM alpine:3.22 AS electoral-data
RUN apk add --no-cache ca-certificates
ARG SNAPSHOT_REV=main
RUN mkdir -p /data && for file in lgas.json wards.json polling-units.csv; do wget -q -T 90 -O "/data/$file" "https://raw.githubusercontent.com/saidiadegoke/nigeria-inec-geo/$SNAPSHOT_REV/data/$file" || exit 1; done

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /dov /app/dov
COPY --from=build /import-inec /app/import-inec
COPY --from=build /app/web /app/web
COPY --from=electoral-data /data /app/data
USER app
ENV GIN_MODE=release
EXPOSE 10000
# Import atomically before starting the public service. Any failure stops startup.
CMD ["/bin/sh","-c","/app/import-inec -dir /app/data && exec /app/dov"]
