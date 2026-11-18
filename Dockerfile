# =========================
# Stage 1: build
# =========================
FROM golang:1.26.5-alpine AS builder

WORKDIR /src

# Dependency layer — di-cache terpisah agar rebuild lebih cepat
# saat hanya source code yang berubah.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO dimatikan agar binary statis (berjalan di image alpine/scratch).
# -trimpath & -s -w mengecilkan ukuran dan menghilangkan path lokal.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/server ./cmd/server

# =========================
# Stage 2: runtime
# =========================
FROM alpine:3.20 AS runtime

# ca-certificates dibutuhkan untuk HTTPS ke Brevo API & PostgreSQL SSL,
# tzdata untuk TimeZone=Asia/Jakarta pada DSN.
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app

ENV TZ=Asia/Jakarta \
    APP_ENV=production \
    GIN_MODE=release \
    APP_PORT=3000

WORKDIR /app

COPY --from=builder /out/server /app/server

USER app

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${APP_PORT}/health" >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/server"]
