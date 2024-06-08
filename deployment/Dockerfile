FROM scratch

ARG TARGETARCH

VOLUME /etc/bastion-web-proxy

COPY build/bastion-web-proxy-linux-${TARGETARCH} /bastion-web-proxy

ENV HOME=/

USER 1001

ENTRYPOINT [ "/bastion-web-proxy" ]

CMD [ "--config-location", "/etc/bastion-web-proxy" ]
