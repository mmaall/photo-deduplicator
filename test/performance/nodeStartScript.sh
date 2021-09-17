#!/bin/bash

# don't forget to source this script
cd /home/ubuntu

sudo apt-get update -y
sudo apt-get upgrade -y

sudo apt-get install awscli -y 
sudo apt-get install systat -y
sudo apt-get install zip -y 


aws s3 cp s3://m2-rocks-db-test/loadTestCode.zip ./

unzip loadTestCode.zip

cd loadTestCode/load_test/

aws s3 cp s3://m2-rocks-db-test/smallSelectedIDs.json ./
aws s3 cp s3://m2-rocks-db-test/mediumSelectedIDs.json ./
aws s3 cp s3://m2-rocks-db-test/largeSelectedIDs.json ./

cd ../../

chown -R ubuntu:ubuntu ./loadTestCode/

dbUrl="localhost/rocks_db_test_db";
username="dbUser"
password="potato"
testType="small"
machineID=0
testLength=2
numThreads=1

