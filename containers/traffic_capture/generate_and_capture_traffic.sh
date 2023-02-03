#!/bin/bash

function teardown() {
    echo "Teardown"

    count=0
    sum=0
    
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
    
    percent=10
    jq -n --arg per "$percent" \
        '{"percent": $per }' > data.json
    # echo "sending data:"
    # echo "download: $avg_download, upload: $avg_upload"

    # echo `cat data.json`
    cat data.json | socat - unix-connect:$SOCKET

    exit
}

function run_sampler() {
    #tcpdump -i eth0 -w /tmp/data.pcap &

    # iperf3 -u -c $HOST -t $RUN_TIME -i 1 -p $PORT -J > temp.txt
    # loss=$(jq '.end.sum.lost_packets' temp.txt)
    # total=$(jq '.end.sum.packets' temp.txt)
    # jq '.end.sum' temp.txt
    # echo "$loss,$total" >> data.txt
    # echo `cat data.txt`

    while [ $SECONDS -lt $END_TIME ]; do
        # iperf3 -u -c $HOST -i 1 -p $PORT -J  > temp.txt
        # loss=$(jq '.end.sum.lost_packets' temp.txt)
        # total=$(jq '.end.sum.packets' temp.txt)
        # jq '.end.sum' temp.txt
        # echo "$loss,$total" >> data.txt
        # echo `cat data.txt`
    done

    exit
}

function kill_all() {
    #killall -s 9 speedtest
    echo "killall called"
    kill -9 `ps aux | grep run_sampler | awk '{print $2}'`
    kill -9 `ps aux | grep iperf | awk '{print $2}'`
}

TOTAL=1
RETRANSMISSION=1
DELAY=5
RUN_TIME=$((DURATION-DELAY))

sleep $DELAY
#rm /tmp/data.pcap /tmp/data.json $SOCKET

trap kill_all SIGTERM

echo "" > data.txt
export -f run_sampler
bash -c run_sampler &
PID=$!
echo "Running run sampler pid: $PID"
echo "waiting for sampler to die"
wait $PID
echo "sampler died"
ps aux


teardown