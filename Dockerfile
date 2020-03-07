FROM golang:1.9

WORKDIR /go/src/github.com/j4y_funabashi/inari-micropub
COPY . .

RUN go build cmd/inari-web/main.go
