FROM golang:1.23.0-alpine AS go-builder
WORKDIR /app
COPY . /app
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && CGO_ENABLED=0 go build -o server .

FROM yimiaoxiehou/alpine
WORKDIR /app
COPY --from=go-builder /app/server /app

EXPOSE 1080
ENTRYPOINT [ "/app/server" ]
