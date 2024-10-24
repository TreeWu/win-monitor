FROM golang:1.20-alpine AS builder
RUN echo "https://mirrors.aliyun.com/alpine/latest-stable/main" > /etc/apk/repositories \
    && echo "https://mirrors.aliyun.com/alpine/latest-stable/community" >> /etc/apk/repositories
RUN apk update && apk add --no-cache git
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY ../go-practise/win_monitor/server /app/win_monitor/server
COPY ./win_monitor/entity.go /app/win_monitor/entity.go
COPY ../go-practise/go.mod /app
RUN ls
RUN go mod tidy
RUN go build -o app ./win_monitor/server/server.go


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
COPY ../go-practise/win_monitor/dashboard/dist /app/win_monitor/dist
RUN apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
CMD ["./app"]

