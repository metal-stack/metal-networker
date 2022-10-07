#!/bin/bash

set -e

validate () {
    echo "----------------------------------------------------------------"
    echo "Validating sample artifacts of metal-networker with ${1}:${2} frr:${3}"
    echo "----------------------------------------------------------------"
    docker build \
        --build-arg OS_NAME="${1}" \
        --build-arg OS_VERSION="${2}" \
        --build-arg FRR_VERSION="${3}" \
        --file Dockerfile.validate \
        . -t metal-networker
}

validate "ubuntu" "20.04" "frr-8"
validate "debian" "10" "frr-8"

# There is no release for jammy available yet
# validate "ubuntu" "22.04" "frr-8"
validate "ubuntu" "22.04" "frr-8"
validate "debian" "11" "frr-8"