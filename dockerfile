FROM golang:1.21.4

WORKDIR /goscrapper

COPY ./ ./
COPY go.mod go.sum ./
RUN go mod download

RUN go build -o goscrapper ./main.go

CMD ./goscrapper