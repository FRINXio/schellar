FROM golang:1.15.3-alpine as builder

RUN mkdir /schellar
WORKDIR /schellar

ADD schellar/. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /go/bin/schellar .

FROM alpine

COPY --from=builder /go/bin/schellar /bin/
ADD startup.sh /

ENV CONDUCTOR_API_URL ''
ENV CHECK_INTERVAL '10'
ENV MONGO_ADDRESS ''
ENV MONGO_USERNAME ''
ENV MONGO_PASSWORD ''

CMD ["sh","startup.sh"]⏎
