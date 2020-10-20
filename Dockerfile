FROM golang:1.15.3-alpine as builder

RUN mkdir /schellar
WORKDIR /schellar

ADD schellar/. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /go/bin/schellar .

FROM alpine

WORKDIR /schellar

COPY --from=builder /go/bin/schellar .
ADD schellar/migrations/ ./migrations

CMD ["./schellar"]
