FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /kube-imds ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /kube-imds /kube-imds
ENTRYPOINT ["/kube-imds"]
