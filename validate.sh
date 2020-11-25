#!/bin/bash

validate () {
    echo "----------------------------------------------------------------"
    echo "Validating sample artifacts of metal-networker with ${1}:${2} frr:${3}"
    echo "----------------------------------------------------------------"
    docker build \
        --no-cache \
        --build-arg OS_NAME="${1}" \
        --build-arg OS_VERSION="${2}" \
        --build-arg FRR_VERSION="${3}" \
        --file Dockerfile.validate \
        . -t metal-networker
}

validate "ubuntu" "20.04" "7.3"
validate "debian" "10" "7.3"

validate "ubuntu" "20.04" "7.5"
validate "debian" "10" "7.5"