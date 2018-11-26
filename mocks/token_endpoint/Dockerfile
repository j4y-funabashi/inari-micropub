FROM golang:1.9

WORKDIR /go/src/github.com/j4y_funabashi/mock-token
COPY main.go .

RUN go get ./...
RUN go install ./...

EXPOSE 80
ENTRYPOINT [ "/go/bin/mock-token" ]
