#!/bin/bash
function teardown() {
    echo "Teardown"

    TOTAL=$(tshark -r /tmp/data.pcap | wc -l)
    RETRANSMISSIONS=$(tshark -r /tmp/data.pcap -Y "tcp.analysis.retransmission" | wc -l)

    rm /tmp/data.pcap

    jq -n --arg total "$TOTAL" --arg ret "$RETRANSMISSIONS" \
        '{"total": $total, "retransmissions": $ret }' > /tmp/data.json

    echo `cat /tmp/data.json`
    cat /tmp/data.json | socat - unix-connect:$SOCKET

    exit
}

function run_sampler() {
    tcpdump -i eth0 -w /tmp/data.pcap &

    while true
    do
        curl -ko /dev/null $URL
        sleep $SLEEP
    done    
}

function kill_all() {
    killall -s 9 curl
    kill -9 `ps aux | grep run_sampler | awk '{print $2}'`
    kill -9 `ps aux | grep tcpdump | awk '{print $2}'`
}

TOTAL=1
RETRANSMISSION=1

rm /tmp/data.pcap /tmp/data.json $SOCKET

trap kill_all SIGTERM

export -f run_sampler
bash -c run_sampler &
PID=$!
echo "Running $PID"
wait $PID

teardown