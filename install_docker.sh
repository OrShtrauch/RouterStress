#!/bin/bash

sudo apt update && sudo apt docker-engine -y
sudo groupadd docker && sudo usermod -aG docker $USER
newgrp docker

sudo systemctl enable docker.service
sudo systemctl enable containerd.service