FROM golang:1.20 as build

WORKDIR /app
COPY src/ /app

RUN go mod download && CGO_ENABLED=0 go build -o book

FROM alpine:latest
COPY --from=build  /app/book /
CMD ["./book"]
