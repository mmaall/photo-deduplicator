#!/bin/bash

numDuplicates=15


# Get the list of files 
files=($(ls -d *.jpg))

# Create the duplicates 
for ((i = 0; i < $numDuplicates; i++));  do

    cp ${files[i]} "duplicate-${files[i]}"

done
