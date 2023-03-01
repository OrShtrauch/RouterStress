#!/bin/sh

SLEEP="$1"

if [ -z "$SLEEP" ]; then
	SLEEP=3
fi

if  [ "$2" != "resume" ]
then
  echo "timestamp,cpu,mem" > /tmp/hardware_data.csv
fi

PREV_TOTAL=0
PREV_IDLE=0

while true
do
	TIMESTAMP=$(date "+%Y-%m-%d-%H:%M:%S")

	MEM_DATA=$(grep Mem  /proc/meminfo | awk '{print $2}')
	
	MEM=$(echo $MEM_DATA | awk '{print (($1-$2)/$1)*100}') #MEM%
	#MEM=$(echo $MEM_DATA | awk '{print ($1-$2)}')

	#CPU=$(sed -n 's/^cpu\s//p' /proc/stat)
	CPU=$(grep "cpu " /proc/stat | awk '{print substr($0,index($0, $2))}')
  	IDLE=$(echo $CPU | awk '{print $3}')
	# change to abosulte cpu usage
	TOTAL=$(echo $CPU | awk '{print $1+$2+$3+$6+$7}')

	DIFF_IDLE=$((IDLE-PREV_IDLE))
	DIFF_TOTAL=$((TOTAL-PREV_TOTAL))

	DIFF_USAGE=$(((1000*(DIFF_TOTAL-DIFF_IDLE)/DIFF_TOTAL+5)/10))

	echo "$TIMESTAMP,$DIFF_USAGE,$MEM" >> /var/tmp/hardware_data.csv
	sleep $SLEEP

	PREV_TOTAL=${TOTAL}
	PREV_IDLE=${IDLE}
done
