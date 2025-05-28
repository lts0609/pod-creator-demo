#!/bin/sh

set -e

if [[ -f "/home/ssh/ssh-password" ]]; then
    SSH_PASSWORD=$(cat /home/ssh/ssh-password)
else
    echo "Not Find SSH_PASSWORD"
    exit 1
fi

echo "root:${SSH_PASSWORD}" | chpasswd

echo "Start SSH Server"
/usr/sbin/sshd -D &

echo "Start Pod-Creator Server"
exec /app/pod-creator