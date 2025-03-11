# Stage 1: Build the binary
FROM golang:1.24.0-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /main ./cmd/main.go

# Stage 2: Minimal runtime image
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /main /main

USER nonroot

EXPOSE 80

ENTRYPOINT ["/main"]
