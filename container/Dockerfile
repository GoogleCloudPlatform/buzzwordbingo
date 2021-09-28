FROM golang:1.16-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY static ./static

RUN go build -o /bingo-services

EXPOSE 8080

CMD [ "/bingo-services" ]