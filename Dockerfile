FROM golang:1.21-alpine as builder

WORKDIR /schellar

COPY schellar/. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /go/bin/schellar .

FROM alpine

ARG git_commit=unspecified
LABEL git_commit="${git_commit}"
LABEL org.opencontainers.image.source="https://github.com/FRINXio/schellar"

WORKDIR /schellar

RUN apk upgrade --available --no-cache
RUN addgroup -S frinx && adduser -S frinx -G frinx

USER frinx

COPY --from=builder /go/bin/schellar .
COPY schellar/migrations/ ./migrations

CMD ["./schellar"]
