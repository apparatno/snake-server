FROM golang:1.12.9-alpine AS builder

WORKDIR /go/src/app

RUN adduser -D -g '' appuser

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o snakeserver .

FROM scratch

WORKDIR "/"

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/app/snakeserver .

USER appuser

EXPOSE 8080

ENTRYPOINT [ "/snakeserver" ]
