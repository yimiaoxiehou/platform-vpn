FROM golang:1.23.0-alpine AS go-builder
WORKDIR /app
COPY . /app
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && CGO_ENABLED=0 go build -o server .

FROM alpine:latest
ENV TZ="Asia/Shanghai"
RUN apk --no-cache --no-progress add \
    ca-certificates \
    tzdata && \
    cp "/usr/share/zoneinfo/$TZ" /etc/localtime && \
    echo "$TZ" >  /etc/timezone
WORKDIR /app
COPY --from=go-builder /app/server /app

EXPOSE 1080
ENTRYPOINT [ "/app/server" ]
