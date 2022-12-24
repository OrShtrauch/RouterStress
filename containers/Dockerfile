FROM python:3.10.6-slim-buster

RUN apt update && apt install -y iputils-ping net-tools
RUN pip3 install --upgrade pip
RUN pip3 install aiohttp asyncio

WORKDIR /var/tmp/stress

COPY scripts/data_handler.py script/data_handler.py
RUN mkdir data/

