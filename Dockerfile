FROM golang:1.15-alpine AS builder

WORKDIR /builder

ADD go.mod .
ADD go.sum .

RUN go mod download

ADD . .

RUN go build -v -o /go/iam .

FROM alpine

RUN apk add tzdata ca-certificates

WORKDIR /go

COPY --from=builder /go/iam /go/iam

RUN adduser -D --uid 1000 iam iam

WORKDIR /go/

RUN mkdir -p /var/log/lab/iam

RUN chown dashboard /var/log/lab/iam -R
RUN chown dashboard /go/iam
RUN ln -sf /usr/share/zoneinfo/America/Fortaleza /etc/localtime

USER iam

ENTRYPOINT /go/iam
