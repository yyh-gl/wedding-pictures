FROM golang:1.25-alpine AS builder

WORKDIR /go/src/github.com/yyh-gl/wedding-pictures

ARG VERSION

ENV TZ="Asia/Tokyo"

COPY . .

RUN apk add --no-cache \
    ca-certificates \
    git \
    make \
    gcc \
    libc-dev \
    curl
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
RUN go install github.com/air-verse/air@latest
RUN go build -o /go/src/github.com/yyh-gl/wedding-pictures/server . 

FROM gcr.io/distroless/base

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/yyh-gl/wedding-pictures/server /app/server
COPY --from=builder /tmp /tmp

EXPOSE 8080

CMD ["/app/server"]
