#!/bin/bash
# This will convert JSON emitted from iostat to CSV
# Currently this only works when the -c and -t flags are used for iostat
# as it only monitors CPU and timestamp is necessary. 
# Holds default values as detailed in the script
# 
# Example: 
#./tocsv.sh -i input.json \
#           -o output.csv
# Arguments
# -i, --input 
#   The input JSON being converted
# -o, --output
#   The output CSV being conveted to

# Default values 
INPUT_FILE="perf-log.json"
CSV_OUTPUT="perf-log.csv"

# Take in arguments 
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do case $1 in
  -i | --input )
    shift; INPUT_FILE=$1
    ;;
  -o | --output )
    shift; CSV_OUTPUT=$1
    ;;  
esac; shift; done
if [[ "$1" == '--' ]]; then shift; fi


# extract statistics array
STAT_LIST=`cat $INPUT_FILE | jq '.sysstat | .hosts | .[0] | .statistics'`

# Are we managing the first row
IS_FIRST_ROW=true

# Metrics being tracked
METRIC_NAMES=('user' 'nice' 'system' 'iowait' 'steal' 'idle') 

# read through statistics array 
for row in $(echo $STAT_LIST | jq -r '.[] | @base64'); do

    # Extract a given row function
    _jq() {
     echo ${row} | base64 --decode | jq -r ${1}
    }
    STAT_OBJECT=$(_jq '.')

    CPU_OBJECT=`echo $STAT_OBJECT | jq '."avg-cpu"'`
    TIMESTAMP=`echo $STAT_OBJECT | jq '.timestamp'`

    # Print the first row if necessary 
    if [ "$IS_FIRST_ROW" = true ]  ; then
        echo -n "timestamp" > $CSV_OUTPUT

        # Write out the metrics being tracked
        for METRIC in ${METRIC_NAMES[@]} ; do
            echo -n ",$METRIC" >> $CSV_OUTPUT 
        done

        IS_FIRST_ROW=false
    fi

    # Write a new line 
    echo >> $CSV_OUTPUT
   
    # Write out the metrics in order
    # this is not guaranteed by iostat
   
    echo -n $TIMESTAMP  >> $CSV_OUTPUT

    for METRIC in ${METRIC_NAMES[@]} ; do
        echo -n "," >> $CSV_OUTPUT
        echo -n $(echo $CPU_OBJECT | jq ".$METRIC") >> $CSV_OUTPUT
    done

done
