FROM docker.io/library/alpine:3.19

RUN apk add --no-cache docker-cli bash

COPY exporter.sh /exporter.sh
RUN chmod +x /exporter.sh

EXPOSE 9487

USER root

CMD ["/exporter.sh"]
