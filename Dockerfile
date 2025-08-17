FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bitbucket-metrics .


FROM ubuntu:noble-20250716

RUN useradd -m -s /bin/bash app_user

COPY --from=builder /app/bitbucket-metrics /usr/local/bin/bitbucket-metrics

USER app_user:app_user

ENTRYPOINT ["/usr/local/bin/bitbucket-metrics"]
