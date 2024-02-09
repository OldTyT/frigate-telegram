FROM arm32v6/golang:1.21.3-alpine3.18 AS builder

WORKDIR $GOPATH/frigate-telegram
COPY . $GOPATH/frigate-telegram

RUN ls -la && exit 1&& go build .

FROM arm32v6/alpine:3.18

RUN apk --no-cache add tzdata

COPY --from=builder /go/frigate-telegram/frigate-telegram /frigate-telegram

USER 1000

ENTRYPOINT ["/frigate-telegram"]
