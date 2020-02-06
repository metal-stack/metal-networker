ARG OS_NAME
ARG OS_VERSION
FROM metalstack/frr:7-${OS_NAME}-${OS_VERSION} AS frr-artifacts

FROM ${OS_NAME}:${OS_VERSION}

ENV FRR_FILES="frr.conf.firewall frr.conf.machine" \
    INTERFACES_FILES="interfaces.firewall interfaces.machine" \
    TESTDATA_DIR="./internal/netconf/testdata"

WORKDIR /tmp
COPY --from=frr-artifacts /artifacts .

RUN ls -alhR /tmp
RUN apt-get update --quiet --quiet \
 && apt-get install \
    --yes \
    --no-install-recommends \
    --quiet \
    --quiet \
    ifupdown2 \
    ./frr_*.deb \
    ./frr-pythontools_*.deb \
    ./libyang*.deb

COPY ${TESTDATA_DIR} /testdata
COPY validate_os.sh /
RUN "/validate_os.sh"