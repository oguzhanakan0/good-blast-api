FROM golang:alpine

ENV PORT=8080

WORKDIR /go/src/
COPY . .
RUN go build .

EXPOSE $PORT

CMD ["go", "run", "main.go"]