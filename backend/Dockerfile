FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/lims-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/lims-migrate ./cmd/migrate

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wkhtmltopdf
WORKDIR /app

COPY --from=builder /build/lims-server .
COPY --from=builder /build/lims-migrate .
COPY --from=builder /build/config ./config
COPY --from=builder /build/templates ./templates

EXPOSE 8080
CMD ["./lims-server"]