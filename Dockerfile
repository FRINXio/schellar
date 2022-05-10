FROM golang:1.15.3-alpine as builder

WORKDIR /schellar

ADD schellar/. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /go/bin/schellar .

FROM alpine:3.14.6

WORKDIR /schellar

RUN addgroup -S frinx && adduser -S frinx -G frinx

USER frinx

COPY --from=builder /go/bin/schellar .
ADD schellar/migrations/ ./migrations

CMD ["./schellar"]
