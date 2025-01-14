#!/bin/bash

set -e

validate () {
    echo "----------------------------------------------------------------"
    echo "Validating sample artifacts of metal-networker with ${1}:${2} frr:${3}"
    echo "----------------------------------------------------------------"
    tag="${1}_${2}_${3}"
    docker build \
        --build-arg OS_NAME="${1}" \
        --build-arg OS_VERSION="${2}" \
        --build-arg FRR_VERSION="${3}" \
        --build-arg FRR_APT_CHANNEL="${4}" \
        --file Dockerfile.validate \
        . -t metal-networker-validate:${tag}

    docker run --interactive \
        --rm \
        --network=none \
        --cap-add=NET_ADMIN \
        --cap-add=NET_RAW \
        --name vali \
        --volume ./pkg/netconf/testdata:/testdata \
        metal-networker-validate:${tag} /validate_os.sh
}

validate "ubuntu" "24.04" "frr-10.2" "noble"
validate "debian" "12" "frr-10.2" "bookworm"