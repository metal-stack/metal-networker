#!/bin/bash

export OS_NAME="ubuntu"
export OS_VERSION="19.04"

validate () {
    echo "----------------------------------------------------------------"
    echo "Validating sample artifacts of metal-networker with ${1}:${2}"
    echo "----------------------------------------------------------------"
    docker build \
        --no-cache \
        --build-arg OS_NAME="${1}" \
        --build-arg OS_VERSION="${2}" \
        . -t metal-networker
}

validate "ubuntu" "19.04"
validate "ubuntu" "19.10"
validate "ubuntu" "20.04"
validate "debian" "10"