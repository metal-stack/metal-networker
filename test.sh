#!/bin/bash

set -eo pipefail

for input in ${FRR_FILES}; do
    echo "/testdata/${input}"
    vtysh --dryrun --inputfile "/testdata/${input}"
done

for input in ${INTERFACES_FILES}; do
    echo "/testdata/${input}"
    test -z $(ifup --syntax-check --all -i "/testdata/${input}" 2>&1 | grep -v "syslogs:")
done
