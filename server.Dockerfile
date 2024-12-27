# 使用 node 镜像作为基础镜像
FROM node:18 AS forntbuild
WORKDIR /app
COPY ./win-monitor-console /app
RUN npm install
RUN npm run build

FROM golang:1.20-alpine AS builder
RUN echo "https://mirrors.aliyun.com/alpine/latest-stable/main" > /etc/apk/repositories \
    && echo "https://mirrors.aliyun.com/alpine/latest-stable/community" >> /etc/apk/repositories
RUN apk update && apk add --no-cache git
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY ./win-monitor-server /app
RUN go mod tidy
RUN go build -o app


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=forntbuild  /app/dist /app/resource/console
RUN apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
CMD ["./app"]

