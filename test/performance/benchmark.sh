#!/bin/bash


# This will run a command and collect metrics of it using iostat. Metrics and 
# output are captured and placed into log files. There are a few arguments
# documented below. All hold default values which are probably not very 
# helpful in your situation 
# 

# Example: 
#./runBenchmark -c "echo hello" \
#    -l command-output.log \
#    -o performance.json

# Arguments 
# -o, --output
#   The file where stdout and stderr will be piped to. 
# -l, --log
#   The file the logs will be shipped to in JSON format. 
#   Follows iostat's format
# -c, --command
#   The binary that is being executed for the perf benchmark.

# Default values
OUTPUT_FILE="output.log"
LOG_FILE="perf-log.json"
COMMAND="echo hello"


# Take in arguments 
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do case $1 in
  -p | --photos )
    shift; PHOTO_DIR=$1
    ;;
  -o | --output )
    shift; OUTPUT_FILE=$1
    ;;
  -l | --log )
    shift; LOG_FILE=$1
    ;; -c | --command ) shift; COMMAND=$1
    ;;
esac; shift; done
if [[ "$1" == '--' ]]; then shift; fi

# Start iostat
# May need to pull the process we spin up

iostat -o JSON -c -t 1 > ${LOG_FILE} &

IOSTAT_PID=$!

# Start the binary command
$COMMAND &> ${OUTPUT_FILE}

# End iostat 
# Send SIGINT to end process gracefully
kill -s SIGINT ${IOSTAT_PID}


