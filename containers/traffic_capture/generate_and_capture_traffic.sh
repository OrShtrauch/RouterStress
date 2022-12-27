#!/bin/bash

function teardown(){
    TOTAL=$(tshark -r data/$FILENAME | wc -l)
    RETRANSMISSIONS=$(tshark -r data/$FILENAME -Y "tcp.analysis.retransmission" | wc -l)

    rm -f data/$FILENAME

    jq --null-input --arg total "$TOTAL" \
    --arg retrnasmissions "$RETRANSMISSIONS" \
    '{"total": $total, "retransmissions": $retrnasmissions }' > data/packet_loss.json

    exit
}

if [ $DURATION == "-1" ]; then
    trap "teardown" SIGTERM
    tcpdump -i eth0 -w data/$FILENAME &
else 
    timeout $(echo $DURATION)s tcpdump -i eth0 -w data/$FILENAME &
fi

START=$SECONDS
TOTAL=1
RETRANSMISSION=1

CURRENT="$((SECONDS-START))"

# echo "start" >> data/test
# echo $SECONDS >> data/test

if [ $DURATION == "-1" ]; then
    while true
    do
        curl -ko /dev/null $URL
        sleep $SLEEP        
    done
else
    while [ "$CURRENT" -lt "$DURATION" ]
    do
        curl -ko /dev/null $URL
        sleep $SLEEP
        CURRENT="$((SECONDS-START))"
        # echo $CURRENT >> data/test
    done    

    # echo "done" >> data/test
    teardown
fi
