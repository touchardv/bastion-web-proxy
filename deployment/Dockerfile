FROM scratch

ARG TARGETARCH
ARG BUILD_DIR=build
ARG BINARY=bastion-web-proxy-linux-${TARGETARCH}

VOLUME /etc/bastion-web-proxy

COPY ${BUILD_DIR}/${BINARY} /bastion-web-proxy

ENV HOME=/

USER 1001

ENTRYPOINT [ "/bastion-web-proxy" ]

CMD [ "--log-level", "debug", "--config-location", "/etc/bastion-web-proxy" ]
