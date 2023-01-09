FROM golang:1.19.4-alpine AS builder
COPY . /dualis-notification
WORKDIR /dualis-notification
ENV GO111MODULE=on
RUN CGO_ENABLED=0 go build -o /main .

FROM alpine:3.17
WORKDIR /
COPY --from=builder /main ./
COPY ./config.env.example ./
COPY ./templates/ ./templates
ENTRYPOINT ["./main"]