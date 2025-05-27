# STEP1
FROM m.daocloud.io/docker.io/golang:1.23-alpine as builder

WORKDIR /usr/local/go/src/

ENV GOPROXY="https://goproxy.cn,direct"

ENV CGO_ENABLED=0

ENV GO111MODULE=on

COPY . ./

RUN go mod tidy && go build -trimpath -ldflags "-s -w" -o pod-creator ./main.go

# STEP2
FROM m.daocloud.io/docker.io/alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

WORKDIR /app

COPY --from=builder /app/pod-creator .

EXPOSE 8080

CMD ["./pod-creator"]