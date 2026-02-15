FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o timetable ./cmd/timetable

FROM alpine:3.19
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /app/timetable .
COPY web/ web/
COPY gtfs/ ./gtfs-seed/

EXPOSE 8080
ENV GTFS_SEED_DIR=/app/gtfs-seed
ENV GTFS_DATA_DIR=/tmp/gtfs
ENV DB_PATH=/tmp/timetable.db
ENV TEMPLATE_DIR=/app/web/templates
ENV STATIC_DIR=/app/web/static

CMD mkdir -p "$GTFS_DATA_DIR" && cp -n "$GTFS_SEED_DIR"/* "$GTFS_DATA_DIR"/ 2>/dev/null; ./timetable
