FROM golang:1.24-alpine AS builder

WORKDIR $GOPATH/frigate-telegram
COPY . $GOPATH/frigate-telegram

RUN apk --no-cache add binutils && \
    go build . && \
    strip frigate-telegram

FROM alpine

RUN apk --no-cache add tzdata

COPY --from=builder /go/frigate-telegram/frigate-telegram /frigate-telegram

USER 1000

ENTRYPOINT ["/frigate-telegram"]
