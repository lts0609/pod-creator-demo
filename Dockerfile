# SETP1
FROM golang:1.23-alpine as stage-bin-build

WORKDIR /workspace

ENV GOPROXY="https://goproxy.cn,direct"

ENV CGO_ENABLED=0

ENV GO111MODULE=on

LABEL stage=stage-bin-build

COPY ./go.mod ./

RUN go mod download

COPY . ./

RUN go build -o pod-creator-demo ./main.go

# SETP2
FROM alpine:3.18

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

WORKDIR /app

COPY --from=builder /app/pod-creator .

EXPOSE 8080

ENTRYPOINT ["./pod-creator"]