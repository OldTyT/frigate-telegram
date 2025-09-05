FROM arm32v6/golang:1.24-alpine AS builder

WORKDIR $GOPATH/frigate-telegram
COPY . $GOPATH/frigate-telegram

RUN apk --no-cache add binutils && \
    go build . && \
    strip frigate-telegram

FROM arm32v6/alpine

RUN apk --no-cache add tzdata

COPY --from=builder /go/frigate-telegram/frigate-telegram /frigate-telegram

USER nobody

ENTRYPOINT ["/frigate-telegram"]
