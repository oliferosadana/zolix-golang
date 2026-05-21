ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/zolix ./cmd/server

FROM scratch

WORKDIR /app

COPY --from=builder /out/zolix /app/zolix
COPY web/static /app/web/static
COPY assets /app/assets

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/zolix"]
