#!/bin/bash

function teardown() {
    echo "Teardown"
    
    # while read line; do
    #     loss=$(echo $line | cut -d ',' -f1)
    #     total=$(echo $line | cut -d ',' -f2)

    #     echo "Total : $total, loss: $loss"

    #     if ! [[ -z "$loss" ]] && ! [[ -z "$total" ]]; then
    #         percent=$((loss / total))

    #         sum=$((sum + percent))
    #         count=$((count + 1))
    #     fi
    # done < <(tr ' ' '\n' < data.txt)

    percent=$(cat data.txt)
    cat data.txt
    
    jq -n --arg loss "$percent" \
        '{"loss": $per }' > data.json
    echo "percent: $percent"
    # echo "sending data:"
    # echo "download: $avg_download, upload: $avg_upload"

    # echo `cat data.json`
    cat data.json | socat - unix-connect:$SOCKET

    exit
}

function run_sampler() {
    #tcpdump -i eth0 -w /tmp/data.pcap &

    iperf3 -u -c $HOST -t $RUN_TIME -i 1 -p $PORT -J > temp.txt

    loss=$(jq '.end.sum.lost_packets' temp.txt)
    total=$(jq '.end.sum.packets' temp.txt)

    percent=$((loss/total))
    echo "percent: $percent" > data.txt

    # jq '.end.sum' temp.txt
    # echo "$loss,$total" >> data.txt
    # echo "1: $RUN_TIME" > /tmp/a.txt
    # while [ "$SECONDS" -lt "$RUN_TIME" ]
    # do
    #     echo "$SECONDS" >> /tmp/a.txt
    #     # iperf3 -u -c $HOST -i 1 -p $PORT -J  > temp.txt
    #     # loss=$(jq '.end.sum.lost_packets' temp.txt)
    #     # total=$(jq '.end.sum.packets' temp.txt)
    #     # jq '.end.sum' temp.txt
    #     # echo "$loss,$total" >> data.txt
    #     # echo `cat data.txt`
    # done

    exit
}

function kill_all() {
    echo "killall called"

    if [ -n "$PID" ]; then
        kill -9 "$PID"
    fi

    killall "iperf3"
}

results_file="/var/tmp/stress/data/results"

TOTAL=1
RETRANSMISSION=1
DELAY=5
RUN_TIME=$((DURATION-DELAY))
echo "runtime is $RUN_TIME"

sleep $DELAY
#rm /tmp/data.pcap /tmp/data.json $SOCKET

trap kill_all SIGTERM

iperf3 -u -c $HOST -t $RUN_TIME -i 1 -p $PORT -J > "$results_file" &
PID=$!
sleep $RUN_TIME

echo "Running run sampler pid: $PID"
echo "waiting for sampler to die"

wait $PID

echo "sampler died"

loss=$(jq '.end.sum.lost_packets' $results_file)
total=$(jq '.end.sum.packets' $results_file)

echo "loss is $loss total is $total" | tee -a /tmp/traffic

percent=$((loss/total))
loss=0
total=0

jq -n --arg loss "$loss" --arg total "$total" \
    '{"loss": $loss, "total": $total }' | socat - unix-connect:$SOCKET
