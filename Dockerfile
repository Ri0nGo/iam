FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go env -w GO111MODULE=on && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/iam cmd\server\main.go


FROM alpine:latest

ARG APP_ROOT=/app/
ARG APP_NAME=iam
ENV RUN=${APP_ROOT}${APP_NAME}
USER root

WORKDIR ${APP_ROOT}

COPY --from=builder ${APP_ROOT}${APP_NAME} ${APP_ROOT}
COPY ./config ${APP_ROOT}config

EXPOSE 8080

CMD ["/bin/sh", "-c", "$RUN"]