FROM golang:1.9

WORKDIR /go/src/github.com/j4y_funabashi/inari-micropub
COPY pkg pkg/
COPY main.go .
COPY Gopkg.lock Gopkg.toml ./

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN go install ./...

EXPOSE 80
ENTRYPOINT [ "/go/bin/inari-micropub" ]
