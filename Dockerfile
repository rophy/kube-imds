FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download
COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /kube-imds ./cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /kube-imds-client ./cmd/client

FROM gcr.io/distroless/static-debian12
COPY --from=builder /kube-imds /kube-imds
COPY --from=builder /kube-imds-client /kube-imds-client
ENTRYPOINT ["/kube-imds"]
