FROM golang:latest

LABEL maintainer="Vadim Inshakov <vadiminshakov@gmail.com>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]