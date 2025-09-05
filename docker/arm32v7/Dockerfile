FROM arm32v7/golang:1.24-alpine AS builder

WORKDIR $GOPATH/frigate-telegram
COPY . $GOPATH/frigate-telegram

RUN apk --no-cache add binutils && \
    go build . && \
    strip frigate-telegram

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /go/frigate-telegram/frigate-telegram /frigate-telegram

ENTRYPOINT ["/frigate-telegram"]
