#!/bin/bash


aws ec2 run-instances \
   --image-id ami-03ba3948f6c37a4b0 \
    --count 1 \
    --instance-type t2.micro \
    --key-name xps-ubuntu \
    --security-group-ids sg-052653a021a5f9e5d \
    --iam-instance-profile Name=ec2-s3ReadWrite\
    --user-data file://nodeStartScript.sh 
