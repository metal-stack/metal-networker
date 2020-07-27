#!/bin/bash

set -eo pipefail

for input in ${FRR_FILES}; do
    echo "/testdata/${input}"
    vtysh --dryrun --inputfile "/testdata/${input}"
done
