# syntax=docker/dockerfile:1

FROM golang:1.22.1

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go test ./tests/test0_test.go

RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/custom-back/ -o ./bin/main

EXPOSE 8080

CMD ["/bin/main"]
