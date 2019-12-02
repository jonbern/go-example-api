FROM golang:1.13-alpine
WORKDIR $GOPATH/src/github.com/jonbern/go-example-api

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["example-api"]