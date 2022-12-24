#!/bin/bash

image_names=( "" "http"  "dns"  "icmp" "port_scan" )
ver="latest"

if [ "$1" == "-h" ]
then
  echo "utility to quickly build docker stress image"

  echo "usgae: ./build_image.sh {name}"
  echo "options 'base', 'http', 'dns',  'icmp', 'port_scan' (leave empty to build all)"

  echo "use flag -w to build parent imgae also"
  echo "example: ./build_image.sh http -w"
  exit 0
fi

# if no image was given, build all of them, with the base image first
if [ "$1" != "" ]
then
  if [ "$1" == "base" ]
  then
    docker build -t stress:$ver -f docker/Dockerfile .
  else    
  # using -w you can build the parent image befroe the given image
    if [ "$2" == "-w" ]
    then 
        docker build -t stress:$ver -f docker/Dockerfile .
    fi    
    docker build -t stress-$1:$ver -f docker/Dockerfile.$1 .
  fi
else
  for name in "${image_names[@]}"
  do
      if [ "$name" != "" ]
      then
        docker build -t stress-$name:$ver -f docker/Dockerfile.$name .
      else
        docker build -t stress:$ver -f docker/Dockerfile .
      fi
  done
fi