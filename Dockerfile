FROM golang:1.10-alpine3.8 as builder

ENV GOOS=linux GOARCH=amd64

WORKDIR /go/src/gitlab.com/pickledrick/g8s-scaler/

COPY main.go /go/src/gitlab.com/pickledrick/g8s-scaler/
COPY g8s /go/src/gitlab.com/pickledrick/g8s-scaler/g8s
COPY Gopkg.* /go/src/gitlab.com/pickledrick/g8s-scaler/

RUN ls -ltr
RUN apk add --no-cache \
        ca-certificates tzdata git curl bash && \
        rm -rf /var/cache/apk/*

RUN curl -s -o /usr/local/bin/dep -L https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod 755 /usr/local/bin/dep

RUN dep ensure
RUN go build

FROM alpine:3.8
RUN apk add --no-cache \
        ca-certificates tzdata && \
        rm -rf /var/cache/apk/*

COPY --from=builder /go/src/gitlab.com/pickledrick/g8s-scaler/g8s-scaler /go/bin/

CMD cd /go/bin/ && ./g8s-scaler