FROM golang:1.8.0
ADD . /go/src/github.com/bouk/extractdata
WORKDIR /go/src/github.com/bouk/extractdata
RUN go get
RUN go install github.com/bouk/extractdata
CMD ["extractdata"]
