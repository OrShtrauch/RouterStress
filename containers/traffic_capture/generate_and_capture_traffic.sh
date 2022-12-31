#!/bin/bash

function teardown(){
    TOTAL=$(tshark -r data/data.pcap | wc -l)
    RETRANSMISSIONS=$(tshark -r data/data.pcap -Y "tcp.analysis.retransmission" | wc -l)

    rm data/data.pcap

    echo "TOTAL: " $TOTAL >> data/test
    echo "RETRANSMISSIONS: " $RETRANSMISSIONS >> data/test
    
    jq -n --arg total "$TOTAL" --arg ret "$RETRANSMISSIONS" \
        '{"total": $total, "retransmissions": $ret }' > data/$FILENAME

    exit
}

if [ $DURATION == "-1" ]; then
    trap "teardown" SIGTERM
    tcpdump -i eth0 -w data/data.pcap &
else 
    timeout $(echo $DURATION)s tcpdump -i eth0 -w data/data.pcap &
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
