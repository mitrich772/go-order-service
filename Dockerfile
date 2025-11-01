# ----------------------
# Stage 1: Build
# ----------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/app
RUN go build -o migrate ./cmd/migrate

# ----------------------
# Stage 2: Runtime
# ----------------------
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .
COPY --from=builder /app/migrate .

COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

#entrypoint
RUN printf '#!/bin/sh\nset -e\n\
echo "🗂 Running migrations..."\n\
./migrate -action up || true\n\
echo "🚀 Starting app..."\n\
exec ./app\n' > /root/entrypoint.sh && chmod +x /root/entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["sh", "/root/entrypoint.sh"]
