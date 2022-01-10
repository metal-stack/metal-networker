#!/bin/bash

testcases="/testdata/frr.conf.*"
for tc in $testcases; do
    print "Testing FRR ${FRR_VERSION} on ${OS_NAME}:${OS_VERSION} with input ${tc}: "
    vtysh --dryrun --inputfile "${tc}"
    if [ $? -eq 0 ]
    then
        printf "\e[32m\xE2\x9C\x94\e[0m\n"
    else
        printf "\e[31m\xE2\x9D\x8C\e[0m\n"
    fi
done
