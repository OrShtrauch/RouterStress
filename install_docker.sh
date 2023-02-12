#!/bin/bash

if [[ $(/usr/bin/id -u) -ne 0 ]]; then
    echo "you must run as root (or with sudo), to run this script"
    exit 1
fi

curl -fsSL https://get.docker.com | sh
groupadd docker && usermod -aG docker $USER
newgrp docker

systemctl enable docker.service
systemctl enable containerd.service