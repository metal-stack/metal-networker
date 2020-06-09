FROM gcr.io/distroless/base

COPY ./bin/metal-networker /

ENTRYPOINT [ "/metal-networker" ]