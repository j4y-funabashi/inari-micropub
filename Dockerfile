FROM golang:1.9

## install dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/github.com/j4y_funabashi/inari-micropub
COPY pkg pkg/
COPY main.go .
COPY Gopkg.lock Gopkg.toml ./

RUN dep ensure
RUN go install ./...

EXPOSE 80
ENTRYPOINT [ "/go/bin/inari-micropub" ]
