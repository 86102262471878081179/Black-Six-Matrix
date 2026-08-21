FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o black-six-matrix ./cmd/omni-proxy/sovereign_allinone.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /app/black-six-matrix .

EXPOSE 8080

ENV SOVEREIGN_MODE=true
ENV PORT=8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./black-six-matrix"]
