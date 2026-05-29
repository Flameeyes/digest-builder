# SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
# SPDX-License-Identifier: 0BSD

# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /digest-builder .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Run as non-root
RUN adduser -D -u 1000 digest
USER digest

COPY --from=builder /digest-builder /usr/local/bin/digest-builder

ENTRYPOINT ["digest-builder"]
