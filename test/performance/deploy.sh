#!/bin/bash


# Default values
IP=3.238.37.52
PHOTO_ZIP=photos.zip
BINARY=photo-deduplicator

# Take in arguments 
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do case $1 in
  -i | --ip )
    shift; IP=$1
    ;;
  -p | --photos )
    shift; PHOTO_ZIP=$1
    ;;
  -b | --binary )
    shift; BINARY=$1
    ;;
  -f | --flag )
    flag=1
    ;;
esac; shift; done
if [[ "$1" == '--' ]]; then shift; fi

echo $IP
echo $PHOTO_ZIP
echo $BINARY