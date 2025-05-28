# STEP1
FROM m.daocloud.io/docker.io/golang:1.23-alpine as builder

WORKDIR /app

ENV GOPROXY="https://goproxy.cn,direct"

ENV CGO_ENABLED=0

ENV GO111MODULE=on

COPY . ./

RUN go mod tidy

RUN go build -trimpath -ldflags "-s -w" -o pod-creator ./main.go

# STEP2
FROM m.daocloud.io/docker.io/alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

RUN apk add --no-cache openssh \
    && sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config \
    && sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config \
    && echo "root:root" | chpasswd \
    && mkdir -p /run/sshd \
    && ssh-keygen -A

WORKDIR /app

COPY --from=builder /app/pod-creator .

EXPOSE 8080 22

CMD ["/bin/sh", "-c", "/usr/sbin/sshd -D & exec \"./pod-creator\""]