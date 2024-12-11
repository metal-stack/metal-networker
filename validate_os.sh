#!/bin/bash

testcases="/testdata/frr.conf.*"
for tc in $testcases; do
    echo -n  "Testing ${FRR_VERSION} on ${OS_NAME}:${OS_VERSION} with input ${tc}: "
    if vtysh --dryrun --inputfile "${tc}";
    then
        printf "\e[32m\xE2\x9C\x94\e[0m\n"
    else
        printf "\e[31m\xE2\x9D\x8C\e[0m\n"
        echo "FRR ${FRR_VERSION} on ${OS_NAME}:${OS_VERSION} produces an invalid configuration"
        exit 1
    fi
done

testcases="/testdata/nftrules*"
for tc in $testcases; do
    echo -n  "Testing nft rules on ${OS_NAME}:${OS_VERSION} with input ${tc}: "
    if nft -c -f "${tc}";
    then
        printf "\e[32m\xE2\x9C\x94\e[0m\n"
    else
        printf "\e[31m\xE2\x9D\x8C\e[0m\n"
        echo "nft input ${tc} on ${OS_NAME}:${OS_VERSION} produces an invalid configuration"
        exit 1
    fi
done