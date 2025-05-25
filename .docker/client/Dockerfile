FROM golang:1.24

WORKDIR /app

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

COPY ./cmd/client ./cmd/client
COPY ./internal/client ./internal/client
COPY ./internal/common ./internal/common

RUN CGO_ENABLED=1 GOOS=linux go build -C ./cmd/client -o client

CMD ["tail", "-f", "/dev/null"]
