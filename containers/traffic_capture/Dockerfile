FROM ubuntu

RUN apt update && apt install iperf3 jq socat -y

WORKDIR /var/tmp/stress
RUN mkdir data/

COPY ./generate_and_capture_traffic.sh script/generate_and_capture_traffic.sh

RUN chmod +x script/generate_and_capture_traffic.sh

ENTRYPOINT ["script/generate_and_capture_traffic.sh"]